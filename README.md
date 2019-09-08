# conjr

---------------------------

## TODO 

- [_] Downlaod Kaltura Installer
- [_] Install Kaltura Classroom
    - Using custom string from configurations
- [_] Run Kaltura classroom
- [_] Close Kaltura classroom after 2 seconds
    - This'll generate localSettings.json to get the resourceID
- [_] Add serial number to google sheet is it isn't already in there
    - Serial Number
    - ResourceID
    - Hostname
    - IP 
    - MAC
- [_] Generate localSettings off of a template instead of just updating resourceID

------------------------------

## Description

Configuring and Reporting for LCC Post-Install.ps1 script. 

This is a script to be run after post installation that ensures installation occured correctly based on a Google Sheet

---------------------------

## Building 

This is meant to build into a windows exe

```sh
~$ env GOOS="windows" go build 
```

--------------------

## Installer format 

```
C:> msiexec.exe /i $file_location /qn /norestart INSTALLDIR="T:\Kaltura\Classroom" ADDLOCAL=ALL KALTURA_RECORDINGS_DIR="T:\Kaltura\Classroom\Recordings\" KALTURA_URL=https://www.kaltura.com/ KALTURA_APPTOKEN=b17a48a9551d8b44076004458c3226a9 KALTURA_APPTOKEN_ID=1_3ma5nq5l KALTURA_PARTNER_ID=1888231 INSTALLDESKTOPSHORTCUT=0  INSTALLPROGRAMSSHORTCUT=1
```
