const puppeteer = require("puppeteer");
const queue = [];
const fs = require("fs");
const filename = "./sitemap.json";
const data = fs.readFileSync(filename);
const details = JSON.parse(data);
const visited = {};
(async () => {
  const browser = await puppeteer.launch({
    headless: false,
    args: ["--no-sandbox", "--disable-setuid-sandbox", "--disable-blink-features=AutomationControlled"],
    dumpio: false
  });
  const page = await browser.newPage();
  await page.evaluate("() =>{Object.defineProperties(navigator,{webdriver:{get: () => false}})}");
  let seedUrl = "http://www.nhc.gov.cn/xcs/yqtb/list_gzbd.shtml";
  const visitPage = async (page, url) => {
    if (visited[url]) {
      return;
    }
    console.log("visiting", url);
    await page.goto(url, {
      waitUntil: "networkidle0"
    });
    await page.waitForSelector(".zxxx_list", {
      visible: true,
      timeout: 3000
    });
    const detailLinks = await page.evaluate(() => {
      const elements = Array.from(document.querySelectorAll(".zxxx_list li"));
      return elements.map((s) => {
        let a = s.getElementsByTagName("a").item(0);
        let txt = a.innerHTML;
        let url = a.getAttribute("href");
        return { title: txt, url: "http://www.nhc.gov.cn" + url };
      });
    });
    let hasDiff = false;
    detailLinks.forEach((x) => {
      if (details[x.url] === undefined) {
        details[x.url] = false;
        hasDiff = true;
      }
    });
    const otherPages = await page.evaluate(() => {
      const elements = Array.from(document.querySelectorAll(".pagination_index_num a"));
      return elements.map((s) => s.getAttribute("href"));
    });
    visited[url] = true;
    if (!hasDiff) {
      return;
    }
    fs.writeFileSync(filename, JSON.stringify(details, null, 2));
    otherPages.forEach((u) => {
      if (visited[u]) {
        return;
      }
      queue.push("http://www.nhc.gov.cn/xcs/yqtb/" + u);
    });
  };
  await visitPage(page, seedUrl);
  while (queue.length > 0) {
    const u = queue.shift();
    await visitPage(page, u);
  }
  console.log(details);
  await browser.close();
})();
