package cmd

import "fmt"

// CompletionCmd generates shell completion scripts.
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell type: bash, zsh, fish." enum:"bash,zsh,fish"`
}

// Run prints the completion script for the given shell.
func (c *CompletionCmd) Run(ctx *Context) error {
	switch c.Shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	}
	return nil
}

const bashCompletion = `# yoy bash completion
_yoy_completions() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="auth mail folders send ls search config version completion"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
        return 0
    fi

    case "${COMP_WORDS[1]}" in
        auth)
            COMPREPLY=($(compgen -W "login logout status" -- "${cur}"))
            ;;
        mail)
            COMPREPLY=($(compgen -W "list search read send reply forward delete move star unstar mark-read mark-unread" -- "${cur}"))
            ;;
        folders)
            COMPREPLY=($(compgen -W "list create delete" -- "${cur}"))
            ;;
        config)
            COMPREPLY=($(compgen -W "get set list path" -- "${cur}"))
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            ;;
    esac
}
complete -F _yoy_completions yoy
`

const zshCompletion = `#compdef yoy

_yoy() {
    local -a commands
    commands=(
        'auth:Manage authentication'
        'mail:Mail operations'
        'folders:Manage mail folders'
        'send:Send an email'
        'ls:List messages'
        'search:Search messages'
        'config:Manage configuration'
        'version:Print version information'
        'completion:Generate shell completions'
    )

    _arguments -C \
        '(-f --folder)'{-f,--folder}'[Mail folder]:folder:' \
        '--json[Output as JSON]' \
        '--plain[Output as plain TSV]' \
        '--color[Color mode]:mode:(auto always never)' \
        '(-v --verbose)'{-v,--verbose}'[Enable verbose output]' \
        '1:command:->cmd' \
        '*::arg:->args'

    case "$state" in
        cmd)
            _describe -t commands 'yoy command' commands
            ;;
    esac
}

_yoy "$@"
`

const fishCompletion = `# yoy fish completion
complete -c yoy -n '__fish_use_subcommand' -a 'auth' -d 'Manage authentication'
complete -c yoy -n '__fish_use_subcommand' -a 'mail' -d 'Mail operations'
complete -c yoy -n '__fish_use_subcommand' -a 'folders' -d 'Manage mail folders'
complete -c yoy -n '__fish_use_subcommand' -a 'send' -d 'Send an email'
complete -c yoy -n '__fish_use_subcommand' -a 'ls' -d 'List messages'
complete -c yoy -n '__fish_use_subcommand' -a 'search' -d 'Search messages'
complete -c yoy -n '__fish_use_subcommand' -a 'config' -d 'Manage configuration'
complete -c yoy -n '__fish_use_subcommand' -a 'version' -d 'Print version information'
complete -c yoy -n '__fish_use_subcommand' -a 'completion' -d 'Generate shell completions'

# auth subcommands
complete -c yoy -n '__fish_seen_subcommand_from auth' -a 'login' -d 'Authenticate via browser'
complete -c yoy -n '__fish_seen_subcommand_from auth' -a 'logout' -d 'Remove stored credentials'
complete -c yoy -n '__fish_seen_subcommand_from auth' -a 'status' -d 'Show auth status'

# mail subcommands
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'list' -d 'List messages'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'search' -d 'Search messages'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'read' -d 'Read a message'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'send' -d 'Send an email'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'reply' -d 'Reply to a message'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'forward' -d 'Forward a message'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'delete' -d 'Delete a message'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'move' -d 'Move a message'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'star' -d 'Star a message'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'unstar' -d 'Unstar a message'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'mark-read' -d 'Mark as read'
complete -c yoy -n '__fish_seen_subcommand_from mail' -a 'mark-unread' -d 'Mark as unread'

# folders subcommands
complete -c yoy -n '__fish_seen_subcommand_from folders' -a 'list' -d 'List folders'
complete -c yoy -n '__fish_seen_subcommand_from folders' -a 'create' -d 'Create a folder'
complete -c yoy -n '__fish_seen_subcommand_from folders' -a 'delete' -d 'Delete a folder'

# config subcommands
complete -c yoy -n '__fish_seen_subcommand_from config' -a 'get' -d 'Get a config value'
complete -c yoy -n '__fish_seen_subcommand_from config' -a 'set' -d 'Set a config value'
complete -c yoy -n '__fish_seen_subcommand_from config' -a 'list' -d 'List all config values'
complete -c yoy -n '__fish_seen_subcommand_from config' -a 'path' -d 'Print config file path'

# Global flags
complete -c yoy -l folder -s f -d 'Mail folder'
complete -c yoy -l json -d 'Output as JSON'
complete -c yoy -l plain -d 'Output as plain TSV'
complete -c yoy -l color -d 'Color mode' -a 'auto always never'
complete -c yoy -l verbose -s v -d 'Enable verbose output'
`
