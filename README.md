# urlscan

Extract URLs from text input to stdin.

Similar to https://pypi.org/project/urlscan/

## Usage

```sh
echo "google.com" >> files.txt
echo "https://stackoverflow.com" >> files.txt
echo "https://github.com" >> files.txt
echo "" >> files.txt
echo "Here's a link to https://example.com in flowing text." >> files.txt
# Without cat...
# urlscan < files.txt
cat files.txt | urlscan
```

This then displays a list of URLs. Enter the number, use tab / shift+tab, or the cursor keys to select an entry, then press enter. The item will then be opened in the browser.

Press q to quit without selecting anything.

```
   [0] google.com
   [1] https://stackoverflow.com
 > [2] https://github.com
   [3] https://example.com
```

