Am I contained
==============

The aim of this code is to detect whether we are running inside of a pid namespace that's not root. It does not check for any other namespaces or cgroups. To do this, permission to load BPF programs is required.

## Build and run

```bash
make && sudo ./main-static
```

## Acknowledgements
Most of this code comes from the libbpfgo repository.
