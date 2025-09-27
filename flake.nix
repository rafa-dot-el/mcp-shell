{
  description = "MCP Shell Server for serving shell AI models";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";  # Pinned version for reproducibility
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          default = pkgs.buildGoModule rec {
            pname = "mcp-shell";
            version = "0.1.0";

            src = ./.;

            vendorHash = "sha256-JFrzxduL0Wr3+CGfAmJbAcaCWRP/vLF6nQWds2aamtw=";

            buildInputs = with pkgs; [
              go
            ];

            meta = with pkgs.lib; {
              description = "MCP Shell Server for serving shell AI models";
              homepage = "https://github.com/rafa-dot-el/mcp-shell";
              license = licenses.gpl3Plus;
              maintainers = with maintainers; [ ];
              platforms = platforms.unix;
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain - latest available in pinned nixpkgs
            go
            gopls
            go-tools
            golangci-lint
            gosec
            goreleaser

            # Development tools
            git
            gh # GitHub CLI
            hugo # Documentation
            go-task # Taskfile runner
            python3Packages.yamllint # YAML validation
            ripgrep # Fast text search
            jq # JSON processing
            curl # HTTP client

            # Node.js and PostCSS for Docsy theme
            nodejs_22
            nodePackages.npm
            nodePackages.postcss
            nodePackages.autoprefixer

            # Security and quality tools
            trivy # Vulnerability scanner
            govulncheck # Go vulnerability checking

            # Debian packaging (if selected)
            

            # Man page generation (if selected)
            
          ];

          shellHook = ''
            # Load project-specific functions and aliases
            if [ -f ./functions.sh ]; then
              source ./functions.sh
            fi

            echo "[+] MCP Shell development environment ready"
            echo "[*] Go version: $(go version)"
            echo "[*] Use 'task' to see available commands"
          '';
        };

        apps = {
          default = flake-utils.lib.mkApp {
            drv = self.packages.${system}.default;
          };
        };
      });
}