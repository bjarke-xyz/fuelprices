import { startOfDay } from "date-fns";
import { createError, getQueryParam, json, parseDate, parseFuelType, parseLanguage } from "../lib/api-utils";
import { PriceRepository } from "../lib/db";
import { Language, getErrorText, getText } from "../lib/localization";
import { Context, Env, FuelType } from "../lib/types";

export const onRequest: PagesFunction<Env> = async (context: Context) => {
    const args = parseArguments(context);
    const priceRepository = new PriceRepository(context.env.DB);
    try {
        const prices = await priceRepository.getPricesForDate(
            args.fuelType,
            args.date
        );
        if (!prices) {
            return createError(getErrorText(args.language));
        }
        return json({
            message: getText(args.language, prices, args.fuelType),
            prices,
        });
    } catch (error) {
        console.error(error);
        return createError(getErrorText(args.language), JSON.stringify(error));
    }
}
interface GetPricesArguments {
    date: Date;
    fuelType: FuelType;
    language: Language;
}
function parseArguments(c: Context): GetPricesArguments {
    return {
        date: parseDate(getQueryParam(c, "now"), startOfDay(new Date())),
        fuelType: parseFuelType(getQueryParam(c, "type") ?? getQueryParam(c, "fueltype")),
        language: parseLanguage(getQueryParam(c, "lang") ?? getQueryParam(c, "language")),
    };
}