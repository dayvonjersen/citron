#!/bin/bash
./zalgo --single=${@%.html}
if [[ $? != 0 ]]; then
    echo -e "\033[31mCompilation failed. Aborted.\033[0m"
    exit 1
fi
js=${@%.html}
js=tmp/$js.js
echo "Linting "$js"..."
acorn --silent $js
if [[ $? -eq 0 ]]; then
    echo -e "\033[32mNo syntax errors.\033[0m"
else
    echo -e "\033[31mSyntax error.\033[0m\nAborted."
    exit 1
fi
echo -e "\033[32mDone\033[0m."
