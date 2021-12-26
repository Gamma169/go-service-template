# go-service-template

Small template for a go service that includes postgres/redis/proxying/mocha tests

All instances of this service are named 'foobar' or 'FOOBAR'


# Rename Service

You should do this immediately after copying the service.  Rename the service to your desired value. For example, if you wanted to call your service `fourbar`, substitute `<service_name>` (and variations, i.e., `<lowercase_service_name>`) with `fourbar`. 

## Rename all instances of foobar in the code
`find ./ -type f -not -path "./.git/*" -not -path "*/node_modules/*" -exec sed -i 's/foobar/<lowercase_service_name>/g' {} \;`

`find ./ -type f -not -path "./.git/*" -not -path "*/node_modules/*" -exec sed -i 's/FOOBAR/<UPPERCASE_SERVICE_NAME>/g' {} \;`

taken from
https://stackoverflow.com/questions/11392478/how-to-replace-a-string-in-multiple-files-in-linux-command-line

## Rename Files/Directories

You also need to rename files and directories.  You can find all the files/directories to be renamed using

`find . -iname "*foobar*"`

You can rename files and directories with:

````for f in `find . -type d -or -type f -iname '*foobar*'`; do mv $f $(echo $f | sed 's/foobar/<service_name>/g'); done````
