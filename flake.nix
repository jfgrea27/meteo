{
  description = "rt-meteo dev environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs { inherit system; config = {
        allowUnfree = true;
      };};
    in {
      devShell = pkgs.mkShell {
        packages = [
          pkgs.uv
          pkgs.go
          pkgs.postgresql
          pkgs.kubectl
          pkgs.kubernetes-helm
          pkgs.just
          pkgs.terraform
          pkgs.tflint
          pkgs.protobuf
          pkgs.grpcurl
          pkgs.pre-commit
        ];
        shellHook = ''
          export GOPATH="$PWD/src/.go"
          export GOROOT="${pkgs.go}/share/go"
          export PATH="$GOPATH/bin:$PATH"
          pre-commit install --allow-missing-config > /dev/null 2>&1
        '';
      };
    });
}
