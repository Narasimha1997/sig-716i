## sig-716i

A CLI tool written in Go that can be used to disrupt wireless connectivity in the area accessible to your wireless interface. This tool scans all the access points and wireless clients in your area and continuously sends large number of [Deauth packets](https://en.wikipedia.org/wiki/Wi-Fi_deauthentication_attack) as per [IEEE 802.11 protocol specification](https://en.wikipedia.org/wiki/IEEE_802.11) which blocks all the wireless devices from accessing the internet via Access Points (AP)

This tool is built for educational purposes, using this tool against wireless defence equipments, medical equipments or public wireless network is strictly not encouraged.

**Disclaimer**: WiFi de-auth attack is illegal in some countries. Using this tool in such countries is not encouraged.

### Requirements:
1. Linux based operating system
2. A system with wireless interface
3. GO programming language tools - [Instructions to install](https://go.dev/doc/install)

### How to install?
1. Clone this repository:
```
git@github.com:Narasimha1997/sig-716i.git
```

2. Go to the project directory and run:
```
sh build.sh
```
If build is successful, it should produce the binary in `bin/`, (`/bin/sig-716i`)

### Running the tool
The tool requires to be run as `sudo` or as a root user.

1. **Run without specifying the wireless interface:** When no wireless interface is selected, the tool looks for the best wireless interface available on the host. If multiple wireless interfaces are present, best one will be selected based on the number of APs discoverable by that interface.

```
sudo ./bin/sig-716i
```

2. **Run with specifying the wireless interface:** You can also specify the interface to use manually. This can be done by passing the option `-i` followed by the name of the wireless interface.
```
sudo ./bin/sig-716i -i <interface-name>
```

Either of the above two commands should start scanning for the wireless APs and devices, later sends the deauth packets to these probed devices. The tool keeps scanning in background so new devices and APs are added to the list as and when they are detected.

3. **Targetting specific APs and Clients:** You can also target a specific interface or devices by specifying them as comma (`,`) seperated strings under `-a` and `-c` options for APs and clients respectively.

```
sudo ./bin/sig-716i -a <BSSID-a>,<BSSID-b>,<BSSID-c>... -c <MAC-a>,<MAC-b>,<MAC-c>....

// example:
sudo ./bin/sig-716i -a 52:54:00:eb:16:9d,02:42:93:53:b4:7b -c 52:54:00:eb:16:9d,02:42:93:53:b4:7b
```

**Tool help**
```
$./bin/sig-716i --help
usage: brute [-h|--help] [-r|--revert] [-i|--iface "<value>"] [-a|--aps
             "<value>"] [-c|--clients "<value>"]

             a CLI tool for bombing wireless networks

Arguments:

  -h  --help     Print help information
  -r  --revert   revert back the host to normal mode of operation. Default:
                 false
  -i  --iface    specify the host interface manually. Default: 
  -a  --aps      list of target BSSIDs of Access Points, ex:
                 52:54:00:eb:16:9d,02:42:93:53:b4:7b,.... Default: 
  -c  --clients  list of MAC addresses of the client devices, ex:
                 52:54:00:eb:16:9d,02:42:93:53:b4:7b,.... Default: 
```

**Notes:**
1. The tool will bring down the wireless interface to monitor mode when starting the attack so you will not be able to access the internet until the tool is running. (you can still use the internet if you have another active wireless interface or ethernet)

2. When exiting, the tool will bring back the interface to normal mode (managed) so you should get internet back, in case it fails, run this command to manually rollback:
```
sudo ./bin/sig-716i --revert -i <interface-name>
```

3. If your wireless interface supports only `2.4Ghz` then it cannot attack devices connected via `5GHz` channel, so it is always recommended to use an interface that supports `5GHz` channel. However this is not mandatory.

4. Though you have a working wireless interface, sometimes, the attack will not be possible because your interface does not support monitoring mode. Monitoring mode has to be supported both by your physical interface as well as the driver. 

**If your physical interface does not support monitoring mode:** You may have to buy a wireless    adapter that supports monitoring mode.

**If your driver does not support monitoring mode:** You may have to patch your wireless driver to support monitoring mode, this is easy on Linux based operating systems. You may have to download the patched driver source code, compile it and build a kernel module, then you can load this kernel module by blacklisting your existing driver. For example the kernel module **r8188eu** has it's patched variant 
[rtl8188eus](https://github.com/aircrack-ng/rtl8188eus).

### TODO:
1. Add exculde lists
2. Provide a cleaner way of passing parameters as TOML config file instead of command line args.

### Credits:
[pywifijammer](https://github.com/DanMcInerney/wifijammer) - python version of wifi jammer

### Contributing:
Feel free to raise issues, send PRs and suggest new features