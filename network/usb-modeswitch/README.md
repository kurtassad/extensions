# usb-modeswitch

Runs [`usb_modeswitch`](https://www.draisberghof.de/usb_modeswitch/) as a system
extension so dual-mode USB devices (common WiFi dongles that first appear as a
fake CD-ROM) are switched into their working mode before the WiFi driver binds.

## Why

Many Realtek USB WLAN adapters enumerate first as mass-storage (`0bda:1a2b`) and
only become a WiFi NIC (e.g. `35bc:0101`) after a SCSI eject / mode-switch.
Stock Talos has no `usb_modeswitch`, so on bare metal the stick can stay stuck
in CD-ROM mode and never produce `wlan0`.

This extension is separate from `wpa-supplicant`: bake both into the image when
you need WiFi on a mode-switching USB adapter.

## Prerequisites

1. Kernel support / driver for the WiFi side of the adapter (e.g. `rtl8852cu`
   extension + `cfg80211`).
2. For association, also install the `wpa-supplicant` extension.

## Installation

See [Installing Extensions](https://github.com/siderolabs/extensions#installing-extensions).
Include this extension in both boot assets (ISO) and the installer if you need
the switch during maintenance / first install.

## Behaviour

On start, `usb-modeswitch-wrapper`:

1. Scans `/sys/bus/usb/devices` for VID:PIDs that have a config under
   `/etc/usb_modeswitch.d/`
2. Runs `usb_modeswitch` for each match
3. Exits `0` if nothing needed switching (already in WiFi mode) or all switches
   succeeded — `restart: untilSuccess` then leaves it alone
4. Exits non-zero on switch failure so Talos retries

The bundled `0bda:1a2b` config targets Realtek sticks that re-enumerate as
`35bc:0101` after `StandardEject`. The full upstream `usb-modeswitch-data`
database is also included for other devices.

## Verify

```bash
# before switch (storage mode)
talosctl --nodes <node> ls /sys/bus/usb/devices/*/idProduct

# after switch + driver
talosctl --nodes <node> service ext-usb-modeswitch status   # STATE: Finished
talosctl --nodes <node> get links | grep wlan
```

## QEMU tip

To exercise this path, pass the dongle through while it is still `0bda:1a2b`
(stop host `usb_modeswitch`/udev from switching it first). Passing through
already-switched `35bc:0101` only tests the driver + `wpa-supplicant`, not this
extension.
