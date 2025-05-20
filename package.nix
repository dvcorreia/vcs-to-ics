{
  lib,
  buildGoModule,
  version ? 0.0 .0,
  commitHash ? "unknown",
}:

buildGoModule (finalAttrs: {
  pname = "vcs-to-ics";
  inherit version;

  src = ./.;

  vendorHash = null;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
    "-X main.commitHash=${commitHash}"
  ];
})
