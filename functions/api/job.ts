import { getHeader, json, statusCode } from "../lib/api-utils";
import { DataFetcher } from "../lib/data-fetcher";
import { PriceRepository } from "../lib/db";
import { Context, Env } from "../lib/types";

export const onRequest: PagesFunction<Env> = async (context: Context) => {
    if (context.request.method !== "POST") {
        return statusCode(400, "must be POST");
    }
    if (context.env.JOB_KEY !== getHeader(context, "Authorization")) {
        return statusCode(401, "missing key")
    }
    await fetchData(context.env);
    return json("job done");
}

async function fetchData(env: Env) {
    const priceRepository = new PriceRepository(env.DB);
    const dataFetcher = new DataFetcher(priceRepository);
    await dataFetcher.fetchData();
}