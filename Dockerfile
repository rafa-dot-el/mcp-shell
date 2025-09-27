# MCP Shell - MCP Shell Server for serving shell AI models
# Copyright (C) 2025 Rafael
# Licensed under GPL-3.0

FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY mcp-shell /usr/local/bin/mcp-shell

# Create a non-root user
USER 65534:65534

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/mcp-shell"]

# Default command
CMD ["--help"]

# Metadata
LABEL org.opencontainers.image.title="MCP Shell"
LABEL org.opencontainers.image.description="MCP Shell Server for serving shell AI models"
LABEL org.opencontainers.image.url="https://github.com/rafa-dot-el/mcp-shell"
LABEL org.opencontainers.image.source="https://github.com/rafa-dot-el/mcp-shell"
LABEL org.opencontainers.image.licenses="GPL-3.0"
LABEL org.opencontainers.image.author="Rafael <<>>"