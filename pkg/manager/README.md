# Bhojpur ISO - Package Manager

The `Bhojpur ISO` features is a multi-platform Package Manager based off from Containers -
it uses Docker (and others) to build packages. It can also version entire `rootfs` and
enables delivery of `OTA`-alike updates, making it a perfect fit for the Edge Computing
and IoT/M2M embedded devices.

It offers a simple `specifications` file format in YAML notation to define both `packages`
and `rootfs`. Based on containers, it can be also used to build stages for Linux from scratch
installations and it can build and track updates for those systems.

## Key Features

- Bhojpur ISO can reuse Gentoo's portage tree hierarchy, and it is heavily inspired from it.
- It builds from containers, but installs, uninstalls and perform upgrades on machines
- Installer doesn't depend on anything ( 0 dep installer !), statically built
- You can install it aside also with your current distro package manager, and start building
and distributing your packages
- Support for packages as "layers"
- It uses SAT solving techniques to solve the deptree
- Support for collections and templated package definitions
- Can be extended with Plugins and Extensions
- Can build packages in Kubernetes (experimental)
- Uses containerd/go-containerregistry to manipulate images - works also daemonless with the img backend
