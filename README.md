# go-get-collibra-backup

# Introduction

This repository is a very simple go application that's intended, at this point, more of a demonstration about how to 
use the Collibra API to pull a backup from the console.  At this point, the project isn't more fully fledged (such as
downloading all backups, cycling backups, or the like.)

# Dependencies and Usage

This project doesn't have any dependencies outside core Go, from a package level.  This project **does** require one 
on disk for the credentials.  As this code is OOTB, that location is `$HOME/.creds/dhc_collibra.json`

The format of this json file is pretty simple, requiring 6 properties to work properly:

```json
{
	"dgc": "https://console-YOUR-ORG.collibra.com",
	"UserName": "ADMINUSER",
	"Password": "ADMINPASS",
	"encryption-key": "ENCRYPTION-KEY-FOR-BACKUP",
	"backup-dir": "/path/to/on/disk/location/",
	"backup-format": "SEE_BELOW"
}
```
# Backup-Format

This script relies on the idea of daily, automated, backups on your instance, and will need to be changed for your environment.

To figure this out, go into your console (dgc link in the JSON), and:
1. Click on "Backups" - Top Bar
2. Click on "Backups" - Left Bar
3. Look at the pattern, like in the screenshot.

![Collibra Console](/docs/CollibraBU.png)

The important part is that for the portion that indicates a date, that you use the placeholder of `<DATE>`.  For example,
lets say you have the following backup item listed:

`Backup_Schedule_catownersunited_2021-12-10_00:00:00`

Your pattern for the backup would then become:

`Backup_Schedule_catownersunited_<DATE>_00:00:00`