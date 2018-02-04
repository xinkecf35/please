go_library(
    name = 'skylark',
    srcs = [
        'builtins.go',
        'parser.go',
        ':builtin_rule_data',
    ],
    deps = [
        '//src/core',
        '//third_party/go:logging',
        '//third_party/go:skylark',
    ],
    visibility = ['//src/parse:all'],
)

genrule(
    name = 'builtin_rules',
    srcs = ['//src/parse:builtin_rules'],
    outs = [
        'c_rules.build_defs',
        'cc_rules.build_defs',
        'go_rules.build_defs',
        'java_rules.build_defs',
        'misc_rules.build_defs',
        'proto_rules.build_defs',
        'python_rules.build_defs',
        'sh_rules.build_defs',
    ],
    cmd = '$TOOL $SRCS',
    tools = [':translate'],
)

filegroup(
    name = 'translate',
    srcs = ['translate.py'],
    binary = True,
)

go_bindata(
    name = 'builtin_rule_data',
    srcs = [':builtin_rules'],
)
