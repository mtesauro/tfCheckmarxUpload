tfCheckmarx-uploader can be used to automatically upload Checkmarx scan result files in XML format to a ThreadFix server.  It is meant to run either as a scheduled job/task or called ad-hoc.  It only produces console output when critical errors are encountered causing an exit with a 1 aka error.

tfCheckmarx-uploader expects files to be named like: 

> {ThreadFix AppID}_[App Name].stuff that follows

So name your files like: 12_[My App]-6.4.2015-17.3.46.xml.  More filename examples are in the log example below.

Instead of being verbose on the console, it creates a log file in a specificed directory.  An example of a run of tfCheckmarx-uploader is:

```
INFO:    2015/04/28 18:11:47 Starting up checkmarx-uploader version 1.3
INFO:    2015/04/28 18:11:47 Opening uploads directory of checkmarx
INFO:    2015/04/28 18:11:47 Reading upload files from checkmarx
INFO:    2015/04/28 18:11:47 Uploading 17 [Example App].Example App-6.4.2015-17.3.46.xml
INFO:    2015/04/28 18:11:53 Successful upload of 17 [Example App].Example App-6.4.2015-17.3.46.xml to ThreadFix AppID 17
INFO:    2015/04/28 18:11:53 Uploading 1717171717 [Bad Filename].Bad Filename-6.4.2015-17.3.46.xml
ERROR:   2015/04/28 18:11:54 checkmarx-uploader.go:286: Error uploading 1717171717 [Bad Filename].Bad Filename-6.4.2015-17.3.46.xml.  Error was: An error occurred during JSON parsing. The error was:   Error parsing Upload response JSON.  Error was: invalid character '<' looking for beginning of value
ERROR:   2015/04/28 18:11:54 checkmarx-uploader.go:289: Parse error for 1717171717 [Bad Filename].Bad Filename-6.4.2015-17.3.46.xml
INFO:    2015/04/28 18:11:54 Problem file moved to checkmarx/parse-errors/1717171717 [Bad Filename].Bad Filename-6.4.2015-17.3.46.xml
INFO:    2015/04/28 18:11:54 Deleting file 17 [Example App].Example App-6.4.2015-17.3.46.xml after uploading to ThreadFix AppID 17
```

To run tfCheckmarx-uploader, you'll first need to create a tfclient.config file with your ThreadFix REST URL + API key and a tfCheckmarx-uploader.config file.

The tfCheckmarx-uploader.config file is used to configure the following:

* watchLocation - a directory / full path / relative path to where the Checkmarx scan files to be uploaded reside
* logLocation - a directory where the output log such as the above example will be written.

tfCheckmarx-uploader follows this workflow:

1. Read config files
1. Setup logging
1. Read all the files in watchLocation as specified in the tfCheckmarx-uploader.config. NOTE: It does not recurse into subdirectories in watchLocation.  It only reads those files in watchLocation.
1. For each file
  1. Parse the file name for an AppID
    * If this fails, move the file to a directory called "parse-errors" in watchLocation.  Create directory if needed.
  1. Upload the scan to the AppID from the filename.
    * If this fails, move the file to a directory called "parse-errors" in watchLocation.  Create directory if needed.
  1. Delete the file after successfully uploading it to ThreadFix.

After a run of tfCheckmarx-uploader, you will have:

* An watchLocation directory with all the successful scan uploads deleted
* Possibly a "parse-errors" directory in watchLocation which includes any file(s) which had errors during handling.

This allows you to run tfCheckmarx-uploader every N minutes without any human interaction.  Occasinally, parse-errors should be checked for problem files but otherwise it should be completely hands off after you setup a scheduled task/cron job.

If you should happen to run tfCheckmarx-uploader without creating a config file, one will be created for you like:

```
$ ./tfCheckmarxUpload 
=====[ Default config file created ]=====

A default configuration file for checkmarx-uploader has been created 
in the current working directory named 'checkmarx-uploader.config'.  Please edit the default
values before running this program again.
Cheers!
=====[ Default config file created ]=====
```
