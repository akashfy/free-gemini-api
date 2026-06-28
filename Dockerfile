FROM akashyadav758/free-gemini-api:latest
COPY gemini-mcp /app/gemini-mcp
RUN chmod +x /app/gemini-mcp
