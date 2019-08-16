# conjr

---------------------------

## TODO 

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

