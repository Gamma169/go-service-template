# go-service-template

Small template for a go service that includes postgres/redis/proxying/mocha tests



# Rename Service

By default, all instances of this service are named 'foobar' or 'FOOBAR'; you should rename this service to your desired service name immediately after copying the service. For example, if you wanted to call your service `fourbear`, substitute `<service_name>` (and variations, i.e., `<lowercase_service_name>`) with `fourbear`. 

## Rename all instances of foobar in the code
```Shell
find ./ -type f -not -path "./.git/*" -not -path "*/node_modules/*" -exec sed -i 's/foobar/<lowercase_service_name>/g' {} \;
```

```Shell
find ./ -type f -not -path "./.git/*" -not -path "*/node_modules/*" -exec sed -i 's/FOOBAR/<UPPERCASE_SERVICE_NAME>/g' {} \;
```

taken from
https://stackoverflow.com/questions/11392478/how-to-replace-a-string-in-multiple-files-in-linux-command-line

**NOTE:**  There are certain places in the code that use models which might have different caps (ex: `FoobarModel`).  Either run the above commands or change manually accordingly.

## Rename Files/Directories

You also need to rename files and directories.  You can find all the files/directories to be renamed using

`find . -iname "*foobar*"`

You can rename files and directories with:

```Shell
for f in `find . -type d -or -type f -iname '*foobar*'`; do mv $f $(echo $f | sed 's/foobar/<service_name>/g'); done
```
