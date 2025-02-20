#!/bin/sh

LIB_DIR="/lib"

# Symlink for x86_64
if [ -e "$LIB_DIR/libwasmvm_muslc.x86_64.a" ]; then
  ln -sf "$LIB_DIR/libwasmvm_muslc.x86_64.a" "$LIB_DIR/libwasmvm.x86_64.a"
  echo "Created symlink: libwasmvm.x86_64.a -> libwasmvm_muslc.x86_64.a"
else
  echo "Source file libwasmvm_muslc.x86_64.a not found, skipping."
fi

# Symlink for aarch64
if [ -e "$LIB_DIR/libwasmvm_muslc.aarch64.a" ]; then
  ln -sf "$LIB_DIR/libwasmvm_muslc.aarch64.a" "$LIB_DIR/libwasmvm.aarch64.a"
  echo "Created symlink: libwasmvm.aarch64.a -> libwasmvm_muslc.aarch64.a"
else
  echo "Source file libwasmvm_muslc.aarch64.a not found, skipping."
fi