package gemini

/*
#include <stdlib.h>
#include <dlfcn.h>

typedef int (*needle_init_fn)(const char* system, const char* tools_json, const char* tool_index_path);
typedef int (*needle_complete_fn)(const char* prompt, int max_new_tokens, char* out_buffer, int buffer_size);
typedef void (*needle_reset_fn)(void);

static void* load_needle_lib(const char* path) {
    return dlopen(path, RTLD_LAZY | RTLD_GLOBAL);
}

static int call_needle_init(void* fn, const char* system, const char* tools_json, const char* tool_index_path) {
    return ((needle_init_fn)fn)(system, tools_json, tool_index_path);
}

static int call_needle_complete(void* fn, const char* prompt, int max_new_tokens, char* out_buffer, int buffer_size) {
    return ((needle_complete_fn)fn)(prompt, max_new_tokens, out_buffer, buffer_size);
}
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"
)

const NeedleEngineVersion = "2.0.3"

type NeedleEngine struct {
	mu        sync.Mutex
	libHandle unsafe.Pointer
	initFn    unsafe.Pointer
	compFn    unsafe.Pointer
	isLoaded  bool
}

var (
	globalNeedle *NeedleEngine
	needleOnce   sync.Once
)

// NeedleResponse matches Needle 2 raw output
type NeedleResponse struct {
	Type          string `json:"type"` // "call" or "respond"
	Success       bool   `json:"success"`
	Reasoning     string `json:"reasoning"`
	Confidence    *float64 `json:"confidence,omitempty"`
	FunctionCalls []struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	} `json:"function_calls"`
	Error     any `json:"error,omitempty"`
	ErrorCode any `json:"error_code,omitempty"`
}

func getLibFileName() string {
	switch runtime.GOOS {
	case "darwin":
		return "libneedle.dylib"
	case "windows":
		return "needle.dll"
	default:
		return "libneedle.so"
	}
}

func getPlatformURL() string {
	base := fmt.Sprintf("https://huggingface.co/Cactus-Compute/needle2/resolve/main/%s", NeedleEngineVersion)
	var platform string
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			platform = "macos-arm64"
		} else {
			platform = "macos-x86_64"
		}
	case "linux":
		if runtime.GOARCH == "arm64" {
			platform = "linux-aarch64"
		} else {
			platform = "linux-x86_64"
		}
	case "windows":
		platform = "windows-x86_64"
	default:
		platform = "macos-arm64"
	}
	return fmt.Sprintf("%s/%s/%s", base, platform, getLibFileName())
}

func ensureNeedleLib() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	cacheDir := filepath.Join(home, ".cache", "cactus-needle", NeedleEngineVersion)
	libPath := filepath.Join(cacheDir, getLibFileName())

	if _, err := os.Stat(libPath); err == nil {
		return libPath, nil
	}

	// Auto-download from Hugging Face if not found
	os.MkdirAll(cacheDir, 0755)
	downloadURL := getPlatformURL()
	fmt.Printf("📥 Auto-downloading Needle 2 Engine (%s) from Hugging Face...\n", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download needle engine: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status downloading needle: %s", resp.Status)
	}

	out, err := os.Create(libPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return libPath, nil
}

func GetNeedleEngine() (*NeedleEngine, error) {
	var initErr error
	needleOnce.Do(func() {
		libPath, err := ensureNeedleLib()
		if err != nil {
			initErr = err
			return
		}

		cPath := C.CString(libPath)
		defer C.free(unsafe.Pointer(cPath))

		handle := C.load_needle_lib(cPath)
		if handle == nil {
			initErr = fmt.Errorf("failed to dlopen needle engine at %s", libPath)
			return
		}

		initSym := C.CString("needle_init")
		compSym := C.CString("needle_complete")
		defer C.free(unsafe.Pointer(initSym))
		defer C.free(unsafe.Pointer(compSym))

		initFn := C.dlsym(handle, initSym)
		compFn := C.dlsym(handle, compSym)

		if initFn == nil || compFn == nil {
			initErr = fmt.Errorf("failed to locate needle C symbols")
			return
		}

		globalNeedle = &NeedleEngine{
			libHandle: handle,
			initFn:    initFn,
			compFn:    compFn,
			isLoaded:  true,
		}
	})

	if initErr != nil {
		return nil, initErr
	}
	return globalNeedle, nil
}

// CompleteTools evaluates user prompt against schemas and returns tool call if triggered
func (n *NeedleEngine) CompleteTools(prompt string, toolsJSON string) (*NeedleResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	cSys := C.CString("")
	cTools := C.CString(toolsJSON)
	defer C.free(unsafe.Pointer(cSys))
	defer C.free(unsafe.Pointer(cTools))

	rc := C.call_needle_init(n.initFn, cSys, cTools, nil)
	if int(rc) < 0 {
		return nil, fmt.Errorf("needle_init error (code %d)", int(rc))
	}

	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	bufSize := 65536
	buf := make([]byte, bufSize)

	res := C.call_needle_complete(n.compFn, cPrompt, C.int(256), (*C.char)(unsafe.Pointer(&buf[0])), C.int(bufSize))
	if int(res) < 0 {
		return nil, fmt.Errorf("needle_complete error (code %d)", int(res))
	}

	jsonStr := C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
	var needleResp NeedleResponse
	if err := json.Unmarshal([]byte(jsonStr), &needleResp); err != nil {
		return nil, fmt.Errorf("failed to parse needle json response: %w", err)
	}

	return &needleResp, nil
}
