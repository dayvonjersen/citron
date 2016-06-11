#!/bin/bash
if [[ $1 = "--quick" ]]; then
    echo "Skipping compilation."
else
    ./zalgo
    if [[ $? != 0 ]]; then
        echo -e "\033[31mCompilation failed. Aborted.\033[0m"
        exit 1
    fi
fi

# . ../rice.sh

cd tmp

echo "Concatenating CSS..."
# ascii_css
#cat ../../sorbet.css/dist/vars.min.css *.css | myth > ../dist/webcomponents.css
echo "/* {{{ */" > tmpcss
# cat ../../sorbet.css/dist/vars.min.css >> tmpcss
tail -n +1 $(ls -d *.css | tr '\n' ' ') | sed 's/==>\(.*\)<==/\/* \}\}\}\n \1 \{\{\{ \*\//' | sed 's/\r//g' >> tmpcss
echo "/* }}} vim:set fdm=marker foldlevel=0: */" >> tmpcss
mv tmpcss ../dist/webcomponents.css
# echo "CSS4 with myth to webcomponents.css"
# myth --no-prefixes tmpcss > 
# # if [[ $? != 0 ]]; then
#     echo "myth failed."
#     echo "aborted"
#     exit 1
# fi
# rm tmpcss

# ascii_js

echo "Concatenating JS..."
cat ../CustomElements.js ../hecomes.js *.js > ../dist/webcomponents.js 

echo -n "Linting JS..."
acorn --silent ../dist/webcomponents.js
if [[ $? -eq 0 ]]; then
    echo -e "\033[32mNo syntax errors.\033[0m"
else
    echo -e "\033[31mSyntax error.\033[0m"
    for js in $(ls *.js); do
        echo -n $js"..."
        acorn --silent $js
        if [[ $? -eq 0 ]]; then
            echo -e "\033[32mNo syntax errors.\033[0m"
        else
            echo -e "\033[31mSyntax error.\033[0m"
        fi
    done
    echo -e "Aborted."
    exit 1
fi
#cd ../dist
#echo "Minifying CSS..."
#minify webcomponents.css > webcomponents.min.css
#echo "Minifying JS..."
#minify webcomponents.js > webcomponents.min.js

echo -e "\033[32mDone\033[0m."
