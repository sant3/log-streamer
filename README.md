## How to dist

### Requirements BE(powershell or unix)

Suggest to install scoop on windows (https://scoop.sh/#/)

* make (scoop install main/make)
* tar 
* openssl (scoop install main/openssl) 

### Requirements FE(powershell or unix)

 * pnpm (tested on 10.24.0)

`make dist-fe` create a **streamer.tar.gz** with client GUI under frontend/dist


`make dist-be` create a multiple tar with binary agent under backed/dist based on OS and ARCH


### TODO
* add BE for x32 ARCH
* finish create automatically self-signed certs for agent
