# Background

Since the beginning of 2020, Covid-19 pandemic caused great changes to people all over the world.

As citizens living in China mainland, we want to find out historical reports about Covid new cases, 

so that we can remember these days, and plan future trips to different provinces of China mainland by looking up their recent Covid case reports.

After some research we could not find any structured historical data or APIs, but there are text reports every day on the Chinese government health site.

# Challenges Facing

to automate the data crawling, we need to bypass the anti-crawling mechanism of the [source website](http://www.nhc.gov.cn/xcs/yqtb/list_gzbd.shtml)
to extract covid new case data from various styles of texts, we need to apply natural language processing techniques.


# What we Achieved

- A mechanism of crawling and parsing data from the source website.
- A structured [JSON data](./docs/data.json) of daily newly confirmed Covid-19 cases reported in China mainland, breaking down by provinces, from February 2022 up to now, updated weekly. We also derived the days without newly confirmed cases for each province, each day.
- A Visualization showing the daily data in the China mainland map with a slider to change the date, colored by case numbers.
- A link to the introduction to Tourist Attractions in the provinces that recently do not have reported new Covid-19 cases.

# Demo URL

Onine demo url: `https://vindurriel.github.io/covid_report/`
