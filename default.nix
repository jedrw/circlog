{ pkgs ? import <nixpkgs> {}, ... }:

pkgs.buildGoModule {
  pname = "circlog";
  version = "0.1.2";
  src = ./.;


  # To update the vendorHash uncomment this line:
  #
  # vendorHash = "sha256:${pkgs.lib.fakeSha256}";
  #
  # And comment out the real hash. Then run nix-build and you will receive an
  # error like this:
  #
  # error: hash mismatch in fixed-output derivation '/nix/store/ml78cpqk0h971canwivanwg9208pcc49-circlog-0.1.2-go-modules.drv':
  #          specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
  #             got:    sha256-IZ8fYN/1oQIS8bFR+AkYcScQXkU75+gOGN2Ft7intes=
  #
  # Take the "got" hash and put that as the value of vendorHash.
  #
  vendorHash = "sha256-IZ8fYN/1oQIS8bFR+AkYcScQXkU75+gOGN2Ft7intes=";
}
