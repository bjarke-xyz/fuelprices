import { parse } from "date-fns";
import { Language } from "./localization";
import { FuelType } from "./types";

export function createErrorObj(
  msg: string,
  error?: unknown
): { message: string; error: string } {
  if (error) {
    console.error("error", error);
  }
  return {
    message: msg,
    error: msg,
  };
}

export function parseDate(
  dateStr: string | undefined | null,
  defaultDate: Date
): Date {
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
