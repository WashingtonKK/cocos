#!/bin/bash
set -e  # Exit immediately if a command fails

# Temporary extraction directory
WORKDIR="/tmp/initramfs-extract"

mkdir -p "$WORKDIR" && \
cp /etc/cocos/rootfs.cpio.gz "$WORKDIR" && \
cd "$WORKDIR" && \
gzip -d < rootfs.cpio.gz | cpio -idmv && \
rm rootfs.cpio.gz && \
rm -f bin/cocos-agent && \
cp /home/washington/cocos/build/cocos-agent bin/ && \
find . | cpio -o -H newc > ../rootfs.cpio && \
gzip ../rootfs.cpio && \
mv ../rootfs.cpio.gz /etc/cocos/test/rootfs.cpio.gz && \
rm -rf "$WORKDIR" && \

echo "✅ Initramfs updated successfully!" && \

/home/cocosai/bin/qemu-svsm/bin/qemu-system-x86_64 \
  -enable-kvm \
  -machine q35 \
  -cpu EPYC \
  -smp 4,maxcpus=16 \
  -m 25G,slots=5,maxmem=30G \
  -netdev user,id=vmnic-c17243ad-e5d2-4d7e-bb50-5e964114da10,hostfwd=tcp::6110-:7002 \
  -device virtio-net-pci,disable-legacy=on,iommu_platform=true,netdev=vmnic-c17243ad-e5d2-4d7e-bb50-5e964114da10,addr=0x2,romfile= \
  -machine confidential-guest-support=sev0-c17243ad-e5d2-4d7e-bb50-5e964114da10,memory-backend=ram1,igvm-cfg=igvm0 \
  -object memory-backend-memfd,id=ram1,size=25G,share=true,prealloc=false \
  -object sev-snp-guest,id=sev0-c17243ad-e5d2-4d7e-bb50-5e964114da10,cbitpos=51,reduced-phys-bits=1 \
  -object igvm-cfg,id=igvm0,file=/etc/cocos/coconut-qemu.igvm \
  -kernel /home/sammy/bzImage.v3.signed \
  -append "quiet console=ttyS0" \
  -initrd /etc/cocos/test/rootfs.cpio.gz \
  -nographic \
  -monitor pty \
  -fsdev local,id=cert_fs,path=/home/washington/Documents/certs1,security_model=mapped \
  -device virtio-9p-pci,fsdev=cert_fs,mount_tag=certs_share \
  -fsdev local,id=env_fs,path=/home/washington/Documents/env,security_model=mapped \
  -device virtio-9p-pci,fsdev=env_fs,mount_tag=env_share