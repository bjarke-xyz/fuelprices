import { LibsqlDialect } from "@libsql/kysely-libsql";
import { addDays, formatISO } from "date-fns";
import { Kysely } from "kysely";
import {
  DayPrices,
  Env,
  FuelType,
  fuelTypeToInteger,
  mapPrice,
  mapPriceDto,
  Price,
} from "./types";

interface Database {
  fuelprices: FuelpricesTable;
}

interface FuelpricesTable {
  fueltype: number;
  date: string;
  price: number;
  prevPrices: string;
}

export class PriceRepository {
  private readonly db: Kysely<Database>;
  constructor(env: Env) {
    this.db = new Kysely<Database>({
      dialect: new LibsqlDialect({
        url: env.LIBSQL_URL,
        authToken: env.LIBSQL_AUTH_TOKEN,
      }),
    });
  }

  async getPricesForDate(
    fuelType: FuelType,
    date: Date
  ): Promise<DayPrices | null> {
    const yesterday = addDays(date, -1);
    const tomorrow = addDays(date, 1);

    const prices = await this.getPricesBetweenDates(
      fuelType,
      yesterday,
      addDays(tomorrow, 1) // Add 1 day so the query becomes correct
    );
    const dayPrices: DayPrices = {
      today: null,
      tomorrow: null,
      yesterday: null,
    };
    if (prices.length === 0) {
      return dayPrices;
    }

    for (let i = 0; i < prices.length; i++) {
      const price = prices[i];
      if (price.date.getDate() === date.getDate()) {
        dayPrices.today = price;
      } else if (price.date.getDate() === yesterday.getDate()) {
        dayPrices.yesterday = price;
      } else if (price.date.getDate() === tomorrow.getDate()) {
        dayPrices.tomorrow = price;
      }
    }
    if (!dayPrices.today) {
      return null;
    }
    return dayPrices;
  }

  async getPricesBetweenDates(
    fuelType: FuelType,
    from: Date,
    to: Date
  ): Promise<Price[]> {
    const result = await this.db
      .selectFrom("fuelprices")
      .selectAll()
      .where("fueltype", "=", fuelTypeToInteger(fuelType))
      .where("date", ">=", formatISO(from))
      .where("date", "<=", formatISO(to))
      .execute();
    // console.log(
    //   "getPricesBetweenDates",
    //   fuelType,
    //   formatISO(from),
    //   formatISO(to),
    //   result
    // );
    const prices = result.map(mapPrice);
    return prices;
  }

  async getPrices(fuelType: FuelType): Promise<Price[]> {
    const result = await this.db
      .selectFrom("fuelprices")
      .selectAll()
      .where("fueltype", "=", fuelTypeToInteger(fuelType))
      .execute();
    const prices = result.map(mapPrice);
    return prices;
  }

  async upsertPrices(prices: Price[]): Promise<void> {
    if (prices.length === 0) {
      return;
    }
    const priceDtos = prices.map(mapPriceDto);
    await this.db
      .insertInto("fuelprices")
      .values(priceDtos)
      .onConflict((oc) =>
        oc.columns(["fueltype", "date"]).doUpdateSet((eb) => {
          return {
            price: eb.ref("excluded.price"),
            prevPrices: eb.ref("excluded.prevPrices"),
          };
        })
      )
      .execute();
  }
}
