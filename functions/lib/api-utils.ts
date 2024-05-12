import { parse } from "date-fns";
import { Language } from "./localization";
import { Context, FuelType } from "./types";

export function createError(msg: string, error?: string): Response {
    if (error) {
        console.error("error", error);
    }
    return json({
        message: msg,
        error: msg
    })
}

export function json(data: unknown): Response {
    return new Response(JSON.stringify(data), {
        headers: {
            'Content-Type': 'application/json'
        }
    })
}

export function statusCode(httpStatusCode: number, msg?: string): Response {
    const body = msg ? JSON.stringify(msg) : null
    return new Response(body, {
        status: httpStatusCode,
        headers: {
            'Content-Type': 'application/json'
        }
    });
}

export function getQueryParam(context: Context, key: string): string | null {
    const url = new URL(context.request.url);
    return url.searchParams.get(key);
}

export function getHeader(context: Context, key: string): string | null {
    return context.request.headers.get(key) ?? null;
}

export function parseDate(dateStr: string | undefined | null, defaultDate: Date): Date {
    let date = defaultDate;
    if (dateStr) {
        try {
            const format = "yyyy-MM-dd";
            date = parse(dateStr, format, new Date());
        } catch (error) {
            console.error(`failed to parse date. input=${dateStr}`, error);
        }
    }
    return date;
}

export function parseLanguage(langStr?: string | null): Language {
    switch (langStr?.toLowerCase()) {
        case "da":
            return "da";
        default:
            return "en";
    }
}

export function parseFuelType(fuelTypeStr?: string | null): FuelType {
    switch (fuelTypeStr?.toLowerCase()) {
        case "octane100":
            return "Octane100";
        case "diesel":
            return "Diesel";
        default:
            return "Unleaded95";
    }
}
