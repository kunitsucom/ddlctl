#!/bin/sh

#ghostplay silent
ghostplay_custom_prompt() { # change prompt
  printf '\e[1;37muser@localhost:~$\e[0m '
}
sh -c "docker kill ddlctl_demo || true" >/dev/null
sh -c "docker run -di --rm --name ddlctl_demo -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=testdb postgres:16" >/dev/null
export DATABASE_DSN="postgres://postgres:password@localhost/testdb?sslmode=disable"
alias bat='bat --paging=never'
alias difft='difft --display=inline'
#ghostplay end

## Example: `ddlctl generate`

#ghostplay silent
sleep 1
echo
ghostplay_custom_prompt
#ghostplay end



# Prepare your annotated model source code
bat /tmp/sample.go

#ghostplay silent
sleep 2
echo
ghostplay_custom_prompt
echo
ghostplay_custom_prompt
#ghostplay end



# Generate DDL
ddlctl generate --dialect postgres --column-tag-go db --ddl-tag-go pgddl --pk-tag-go pk --src /tmp/sample.go --dst /tmp/sample.sql

#ghostplay silent
sleep 2
echo
ghostplay_custom_prompt
echo
ghostplay_custom_prompt
#ghostplay end



# Check generated DDL file
bat /tmp/sample.sql

#ghostplay silent
sleep 2
echo
ghostplay_custom_prompt
echo
ghostplay_custom_prompt
#ghostplay end



## Example: `ddlctl diff` and `ddlctl apply`

#ghostplay silent
sleep 1
echo
ghostplay_custom_prompt
#ghostplay end



# Apply DDL
ddlctl apply --dialect postgres "${DATABASE_DSN}" /tmp/sample.sql --auto-approve

#ghostplay silent
sleep 2
echo
ghostplay_custom_prompt
echo
ghostplay_custom_prompt
#ghostplay end



# Edit DDL and check diff
diff -uw /tmp/sample.sql /tmp/diff.sql | bat --language diff

#ghostplay silent
sleep 2
echo
ghostplay_custom_prompt
echo
ghostplay_custom_prompt
#ghostplay end



# Diff current database schema and DDL file
ddlctl diff --dialect postgres "${DATABASE_DSN}" /tmp/diff.sql | bat --language sql

#ghostplay silent
sleep 2
echo
ghostplay_custom_prompt
echo
ghostplay_custom_prompt
#ghostplay end



# Apply DDL
ddlctl apply --dialect postgres "${DATABASE_DSN}" /tmp/diff.sql --auto-approve

#ghostplay silent
sleep 4
#ghostplay end
