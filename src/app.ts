import { addDays, startOfDay } from "date-fns";
import { Hono } from "hono";
import { serveStatic } from "hono/bun";
import { logger } from "hono/logger";
import {
  createErrorObj,
  parseDate,
  parseFuelType,
  parseLanguage,
} from "./lib/api-utils";
import { DataFetcher } from "./lib/data-fetcher";
import { PriceRepository } from "./lib/db";
import { getErrorText, getText, Language } from "./lib/localization";
import { FuelType, getEnv } from "./lib/types";

const app = new Hono();
app.use(logger());

app.post("/api/job", async (c) => {
  const env = getEnv();
  if (env.JOB_KEY !== c.req.header("Authorization")) {
    return c.text("missing key", 401);
  }
  const priceRepository = new PriceRepository(env);
  const dataFetcher = new DataFetcher(priceRepository);
  await dataFetcher.fetchData();
  return c.text("success", 200);
});

app.get("/api/price-trends", async (c) => {
  const env = getEnv();
  const from = parseDate(
    c.req.query("from"),
    startOfDay(addDays(new Date(), -1))
  );
  const to = parseDate(c.req.query("to"), startOfDay(new Date()));
  const fuelType = parseFuelType(c.req.query("type"));
  const priceRepository = new PriceRepository(env);
  try {
    const prices = await priceRepository.getPricesBetweenDates(
      fuelType,
      from,
      to
    );
    if (!prices) {
      return c.json(createErrorObj("could not get prices"), 500);
    }
    const priceResult: {
      dates: Date[];
      prices: number[];
    } = { dates: [], prices: [] };
    for (const price of prices) {
      priceResult.dates.push(price.date);
      priceResult.prices.push(price.price);
    }
    return c.json(priceResult);
  } catch (error) {
    return c.json(createErrorObj("get all prices failed", error), 500);
  }
});

app.get("/api/prices", async (c) => {
  const env = getEnv();
  const args: {
    date: Date;
    fuelType: FuelType;
    language: Language;
  } = {
    date: parseDate(c.req.query("now"), startOfDay(new Date())),
    fuelType: parseFuelType(c.req.query("type") ?? c.req.query("fueltype")),
    language: parseLanguage(c.req.query("lang") ?? c.req.query("language")),
  };
  const priceRepository = new PriceRepository(env);
  try {
    const prices = await priceRepository.getPricesForDate(
      args.fuelType,
      args.date
    );
    if (!prices) {
      return c.json(createErrorObj(getErrorText(args.language)));
    }
    return c.json({
      message: getText(args.language, prices, args.fuelType),
      prices,
    });
  } catch (error) {
    return c.json(createErrorObj(getErrorText(args.language), error));
  }
});

app.get("*", serveStatic({ root: "./dist" }));

export default app;
