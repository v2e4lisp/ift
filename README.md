## Usage

```
ift [-d dir] [-wait] [-watchfile path] [-n interval] [-p patterns] [-hidden] command

OPTIONS:
  -d=".": Watch directory
  -hidden=false: Watch hidden file
  -n=2s: Interval between command execution
  -p="": Specify file name patterns to watch. Multiple patterns should be seperated by comma. If pattern is not specified, all files in the dir will be watched(except hidden files). You can also use watch file to specify patterns.
  -wait=false: Wait for last command to finish.
  -watchfile=".watch": Watch file contains file name patterns. ift use these patterns to determine which files to watch. If watchfile is not specified, ift will try to load .watch file under the watch directory. You can also specify patterns using -p option. 
``
