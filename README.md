# go-service-template

Small template for a go service that includes postgres/redis/proxying/mocha tests

All instances of this service are named 'foobar' or 'FOOBAR'


# Rename Service

You should do this immediately after copying the service.  Rename the service to your desired value

## Rename all instances of foobar in the code
`find ./ -type f -not -path "./.git/*" -not -path "*/node_modules/*" -exec sed -i 's/foobar/<lowercase_service_name>/g' {} \;`

`find ./ -type f -not -path "./.git/*" -not -path "*/node_modules/*" -exec sed -i 's/FOOBAR/<UPPERCASE_SERVICE_NAME>/g' {} \;`

taken from
https://stackoverflow.com/questions/11392478/how-to-replace-a-string-in-multiple-files-in-linux-command-line

## Rename Files/Directories

You also need to rename files and directories.  You can find all the files/directories to be renamed using

`find . -iname "*foobar*"`

**TODO:** For now need to rename them manually.  Might look into a command like the above to do all of them

