{
  description = "Optional nix flake for the ns-ci tool";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs";

  outputs =
    { self, nixpkgs, ... }@inputs:
    let
      name = "ns-ci";

      supportedSystems = [
        "x86_64-linux"
        "aarch64-darwin"
        "aarch64-linux"
      ];

      forEachSupportedSystem =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            pkgs = import nixpkgs { inherit system; };
            system = system;
          }
        );

      vendorHash = nixpkgs.lib.fakeHash;
    in
    {
      packages = forEachSupportedSystem (
        { pkgs, system }:
        {
          default = pkgs.buildGo124Module {
            inherit name vendorHash;
            src = ./.;
            doCheck = true;
          };
          format = pkgs.writeShellApplication {
            name = "format";
            runtimeInputs = [
              pkgs.gci
            ];
            text = ''
              gci write -s standard -s default -s localmodule ./
            '';
          };
          checks = pkgs.writeShellApplication {
            name = "checks";
            runtimeInputs = [
              pkgs.actionlint
              pkgs.golangci-lint
              pkgs.typos
            ];
            text = ''
              set +e
              echo "GolangCI Lint:"
              golangci-lint run
              echo "Action Lint:"
              actionlint
              echo "Typos:"
              typos && echo "0 issues."
            '';
          };
          test-release = pkgs.writeShellApplication {
            name = "test-release";
            runtimeInputs = [
              pkgs.goreleaser
            ];
            text = "goreleaser release --snapshot --clean";
          };
        }
      );

      devShells = forEachSupportedSystem (
        { pkgs, system }:
        {
          default = pkgs.mkShell {
            packages = [
              # Development dependencies
              pkgs.cobra-cli
              pkgs.go_1_24
              pkgs.gopls
              pkgs.oapi-codegen

              # Linting helpers
              pkgs.actionlint
              pkgs.gci
              pkgs.golangci-lint
              pkgs.typos
              pkgs.gotestsum
              pkgs.markdownlint-cli
              pkgs.gocover-cobertura
              pkgs.goreleaser

              # Shell script helpers
              self.packages.${system}.checks
              self.packages.${system}.format
              self.packages.${system}.test-release

              # Other
              pkgs.nixpkgs-fmt
            ];
          };
        }
      );
    };
}
