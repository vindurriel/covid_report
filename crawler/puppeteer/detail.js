const puppeteer = require("puppeteer");
const queue = [];
const fs = require("fs");
const filename = "./sitemap.json";
const data = fs.readFileSync(filename);
const details = JSON.parse(data);
const pageIdRegex = /\/([a-z0-9]+)\.shtml/;
(async () => {
  const browser = await puppeteer.launch({
    headless: false,
    args: ["--no-sandbox", "--disable-setuid-sandbox", "--disable-blink-features=AutomationControlled"],
    dumpio: false
  });
  const page = await browser.newPage();
  await page.evaluate("() =>{Object.defineProperties(navigator,{webdriver:{get: () => false}})}");
  const visitPage = async (page, url) => {
    const pageId = url.match(pageIdRegex)[1];
    console.log("visiting", pageId);
    await page.goto(url, {
      waitUntil: "networkidle0"
    });
    await page.waitForSelector(".con", {
      visible: true,
      timeout: 3000
    });
    const paragraphs = await page.evaluate(() => {
      const elements = Array.from(document.querySelectorAll(".con p"));
      return elements
        .map((p) => {
          return p.textContent.trim();
        })
        .filter((x) => x !== "");
    });
    details[url] = true;
    fs.writeFileSync("data/" + pageId, paragraphs.join("\n"));
    fs.writeFileSync(filename, JSON.stringify(details, null, 2));
  };
  const urls = Object.keys(details).filter((k) => details[k] === false);
  for (let i = 0; i < urls.length; i++) {
    await visitPage(page, urls[i]);
  }
  await browser.close();
})();
