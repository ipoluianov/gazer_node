@ECHO OFF

ECHO package system > system/system/version.go
ECHO. >> system/system/version.go
ECHO const BUILDTIME = "%time% %date%" >> system/system/version.go
ECHO const VERSION = "2.4.4" >> system/system/version.go
