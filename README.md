## Usage

```
Usage:
  ift [-d dir] [-ignorefile path] [-n interval] [-p patterns] [-wait] [-hidden] command

Options:
  -d=".": Watch directory
  -hidden=false: Watch hidden file
  -ignorefile=".iftignore": contains file patterns to ignore. ift use these patterns to determine which files to ignore. If ignorefile is not specified, ift will try to load .iftignore file under the watch directory. You can also specify patterns using -p option. 
  -n=2s: Interval between command execution
  -p="": Specify file name patterns to ignore. Multiple patterns should be seperated by comma. If pattern is not specified, all files in the dir will be watched(except hidden files). You can also use ignore file to specify patterns.
  -wait=false: Wait for last command to finish.
```
