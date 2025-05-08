{
  description = "Optional nix flake for the ns-ci tool";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs";

  outputs = { self, nixpkgs, ... }@inputs:
    let
      name = "ns-ci";

      supportedSystems = [ "x86_64-linux" "aarch64-darwin" "aarch64-linux" ];

      forEachSupportedSystem = f:
        nixpkgs.lib.genAttrs supportedSystems
        (system: f { pkgs = import nixpkgs { inherit system; }; });

      vendorHash = nixpkgs.lib.fakeHash;
    in {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go_1_24
            pkgs.gopls
            pkgs.golangci-lint
            pkgs.gci

            pkgs.oapi-codegen
            pkgs.cobra-cli

            pkgs.nixpkgs-fmt
          ];
        };
      });

      packages = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.buildGo124Module {
          inherit name vendorHash;
          src = ./.;
          doCheck = true;
        };

      });
    };
}
