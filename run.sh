set -e
cd crawler/puppeteer/
node list.js
node detail.js
cd -
go run .
git add crawler/puppeteer/ docs/
git status
