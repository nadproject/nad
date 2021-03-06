#/usr/bin/env bash
__nad_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# commands are the valid commands
commands=("add" "view" "edit" "remove" "find"  "sync" "login" "logout" "help" "version")

_complete_root_command() {
    COMPREPLY=($(compgen -W "${commands[*]}" "${current_word}"))
}

_complete_view_command() {
    names=$(nad view --name-only)
    query=${COMP_WORDS[*]:2}

    __nad_debug "${FUNCNAME[0]} query: ${query}"

    while read -r line; do
        if [[ ${line} == *"${query}"* ]]; then
            COMPREPLY+=("${line}")
        fi
    done <<< "$names"
}

_nad_completions() {
    local current_word="${COMP_WORDS[${COMP_CWORD}]}"

    __nad_debug "COMP_WORDS: ${COMP_WORDS[*]} COMP_CWORD: ${COMP_CWORD} current_word: ${current_word}"

    if [[ "${COMP_CWORD}" -eq 1 ]]; then
        _complete_root_command
    elif [[ "${COMP_CWORD}" -ge 2 ]]; then
        cmd=${COMP_WORDS[1]}

        __nad_debug "cmd: ${cmd}"

        if [[ ( "${cmd}" == view ) || ( "${cmd}" == v ) || ( "${cmd}" == add ) || ( "${cmd}" == a ) ]]; then
            _complete_view_command
        fi
    fi

}

complete -F _nad_completions nad
