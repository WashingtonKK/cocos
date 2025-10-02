#!/usr/bin/env bash
set -euo pipefail

# --- Config ---
SRC_INITRAMFS="/etc/cocos/rootfs.cpio.gz"
DEST_DIR="/etc/cocos/test"
DEST_INITRAMFS="$DEST_DIR/rootfs.cpio.gz"
NEW_AGENT="/home/washington/cocos/build/cocos-agent"
QEMU="/home/cocosai/bin/qemu-svsm/bin/qemu-system-x86_64"

KERNEL="/home/sammy/bzImage.v3.signed"
CERTS_FS_PATH="/home/washington/Documents/certs1"
ENV_FS_PATH="/home/washington/Documents/env"
HOSTFWD_PORT="6110"
GUEST_GRPC_PORT="7002"

# --- Guards ---
[ -f "$SRC_INITRAMFS" ] || { echo "Missing: $SRC_INITRAMFS"; exit 1; }
[ -f "$NEW_AGENT" ]     || { echo "Missing: $NEW_AGENT"; exit 1; }
[ -x "$QEMU" ]          || { echo "Missing/exec: $QEMU"; exit 1; }
[ -f "$KERNEL" ]        || { echo "Missing: $KERNEL"; exit 1; }
mkdir -p "$DEST_DIR"

# --- Workdir & cleanup ---
WORKDIR="$(mktemp -d -t initramfs-XXXXXX)"
cleanup() { rm -rf "$WORKDIR"; }
trap cleanup EXIT

echo "→ Working in $WORKDIR"
cp "$SRC_INITRAMFS" "$WORKDIR/rootfs.cpio.gz"

# --- Unpack ---
cd "$WORKDIR"
gzip -dc rootfs.cpio.gz | cpio -idmv
rm -f rootfs.cpio.gz

# --- Replace agent ---
install -m 0755 "$NEW_AGENT" "$WORKDIR/bin/cocos-agent"

# --- Repack ---
( find . | cpio -o -H newc ) | gzip -9 > "$DEST_INITRAMFS"

echo "✅ Initramfs written to $DEST_INITRAMFS"

# --- Run QEMU (skip with: ./update-initramfs.sh --no-run) ---
if [[ "${1:-}" != "--no-run" ]]; then
  echo "▶ Launching QEMU…"
  "$QEMU" \
    -enable-kvm \
    -machine q35 \
    -cpu EPYC \
    -smp 4,maxcpus=16 \
    -m 25G,slots=5,maxmem=30G \
    -netdev user,id=vmnic-c17243ad-e5d2-4d7e-bb50-5e964114da10,hostfwd=tcp::${HOSTFWD_PORT}-:${GUEST_GRPC_PORT} \
    -device virtio-net-pci,disable-legacy=on,iommu_platform=true,netdev=vmnic-c17243ad-e5d2-4d7e-bb50-5e964114da10,addr=0x2,romfile= \
    -machine confidential-guest-support=sev0-c17243ad-e5d2-4d7e-bb50-5e964114da10,memory-backend=ram1,igvm-cfg=igvm0 \
    -object memory-backend-memfd,id=ram1,size=25G,share=true,prealloc=false \
    -object sev-snp-guest,id=sev0-c17243ad-e5d2-4d7e-bb50-5e964114da10,cbitpos=51,reduced-phys-bits=1 \
    -object igvm-cfg,id=igvm0,file=/etc/cocos/coconut-qemu.igvm \
    -kernel "$KERNEL" \
    -append "quiet console=ttyS0" \
    -initrd "$DEST_INITRAMFS" \
    -nographic \
    -monitor pty \
    -fsdev local,id=cert_fs,path="$CERTS_FS_PATH",security_model=mapped \
    -device virtio-9p-pci,fsdev=cert_fs,mount_tag=certs_share \
    -fsdev local,id=env_fs,path="$ENV_FS_PATH",security_model=mapped \
    -device virtio-9p-pci,fsdev=env_fs,mount_tag=env_share
fi
