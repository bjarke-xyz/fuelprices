import { addDays, startOfDay } from "date-fns";
import { createError, getQueryParam, json, parseDate, parseFuelType } from "../lib/api-utils";
import { Context, Env } from "../lib/types";
import { PriceRepository } from "../lib/db";

export const onRequest: PagesFunction<Env> = async (context: Context) => {
    const from = parseDate(
        getQueryParam(context, "from"),
        startOfDay(addDays(new Date(), -1))
    );
    const to = parseDate(getQueryParam(context, "to"), startOfDay(new Date()));
    const fuelType = parseFuelType(getQueryParam(context, "type"));
    const priceRepository = new PriceRepository(context.env.DB);
    try {
        const prices = await priceRepository.getPricesBetweenDates(
            fuelType,
            from,
            to
        );
        if (!prices) {
            return createError("could not get prices");
        }
        const priceResult: {
            dates: Date[];
            prices: number[];
        } = { dates: [], prices: [] }
        for (const price of prices) {
            priceResult.dates.push(price.date);
            priceResult.prices.push(price.price);
        }
        return json(priceResult);
    } catch (error) {
        console.error(error);
        return createError("get all prices failed", JSON.stringify(error));
    }
}
