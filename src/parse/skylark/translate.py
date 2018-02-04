#!/usr/bin/python3
"""Small script to translate current built-in rules to Skylark's dialect.

This is, needless to say, pretty crude, but serves while I'm
experimenting with Skylark.
"""

import os, re, sys


# Remove all the PEP-484 style type hints.
PEP484_RE = re.compile(r':(?:(?:bool|int|str|list|dict|function)\|?)+([,)=])')
# "raise" is not supported as a statement. We will try supplying it as a builtin...
RAISE_RE = re.compile('raise [A-Za-z]+(.*)\n')


def rewrite(src):
    with open(src) as f:
        contents = f.read()
    contents = PEP484_RE.sub(r'\1', contents)
    contents = RAISE_RE.sub('fail(\\1)\n', contents)
    # Skylark doesn't support the 'is' operator, but '== None' is close enough.
    contents = contents.replace(' is None', ' == None')
    with open(os.path.basename(src), 'w') as f:
        f.write(contents)


if __name__ == '__main__':
    for src in sys.argv[1:]:
        rewrite(src)
