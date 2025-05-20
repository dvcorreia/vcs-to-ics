{
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      supportedSystems = [
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
        "x86_64-linux"
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);
      nixpkgsFor = forAllSystems (
        system:
        import nixpkgs {
          inherit system;
          overlays = [ self.overlay ];
        }
      );

      version = self.shortRev or self.dirtyShortRev;
      commitHash = self.rev or self.dirtyRev;
    in
    {
      overlay = final: _: {
        vcs-to-ics = final.callPackage ./package.nix {
          inherit version commitHash;
        };
      };

      formatter = forAllSystems (system: (nixpkgsFor.${system}).nixfmt-tree);

      packages = forAllSystems (system: {
        default = (nixpkgsFor.${system}).vcs-to-ics;
        vcs-to-ics = (nixpkgsFor.${system}).vcs-to-ics;
      });

      devShells = forAllSystems (
        system: with nixpkgsFor.${system}; {
          default = mkShell {
            inputsFrom = [ vcs-to-ics ];
            packages = [
              git
              copywrite
            ];
          };
        }
      );
    };
}
