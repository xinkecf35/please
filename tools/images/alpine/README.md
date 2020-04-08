Alpine Linux image
------------------

This tests Please on CI in an Alpine environment which can differ in some ways
from the standard Ubuntu setup.

Note that the glibc-compat package must be added to an Alpine environment to
use standard upstream Please builds. It is not needed here since this Please
is built entirely within Alpine.

Some packages have different versions in arbitrary ways - for example Python,
Go and Java. Depending on what's available they may be newer or older than the
Ubuntu image (as a rule Go tends to be a bit older since we get the latest one
in the standard image since it builds releases, whereas Python and Java often
have newer packages available here).
