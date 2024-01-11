#!/usr/bin/env bash

  echo '	// START CASES DO NOT EDIT'
  echo '	switch token {'
  grep -E "^\tTOKEN_[A-Za-z0-9_]+ +TokenType += +[\"\`][A-Za-z0-9_]+[\"\`]" "${1:?}" | while read -r LINE; do
    const=$(awk '{print $1}' <<<"${LINE:-}")
    literal=$(awk '{print $4}' <<<"${LINE:-}")
    case "${literal:?}" in
      '"IDENT"')
        echo -e "\tdefault:"
        echo -e "\t\treturn ${const:?}"
        ;;
      '"OPEN_PAREN"' | '"CLOSE_PAREN"' | '"COMMA"' | '"SEMICOLON"' | '"ILLEGAL"' | '"EOF"')
        continue
        ;;
      *)
        echo -e "\tcase ${literal:?}:"
        echo -e "\t\treturn ${const:?}"
        ;;
    esac
  done
  echo '	}'
  echo '	// END CASES DO NOT EDIT'
