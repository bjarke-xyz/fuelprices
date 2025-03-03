import { formatISO, parseISO } from "date-fns";

export type Env = {
  LIBSQL_URL: string;
  LIBSQL_AUTH_TOKEN: string;
  JOB_KEY: string;
};

export type FuelType = "Unleaded95" | "Octane100" | "Diesel";

export function fuelTypeToOkItemNumber(f: FuelType): number {
  switch (f) {
    case "Octane100":
      return 533;
    case "Diesel":
      return 231;
    default:
      return 536;
  }
}

export function fuelTypeToInteger(f: FuelType): number {
  switch (f) {
    case "Unleaded95":
      return 1;
    case "Octane100":
      return 2;
    case "Diesel":
      return 3;
  }
}

export function integerToFuelType(i: number): FuelType {
  switch (i) {
    case 1:
      return "Unleaded95";
    case 2:
      return "Octane100";
    case 3:
      return "Diesel";
    default:
      throw new Error(`Unknown integer fuel type '${i}'`);
  }
}

export interface PreviousPrice {
  detectionTimestamp: Date;
  price: number;
}
export function mapPrevPriceDto(prevPrice: PreviousPrice): PreviousPriceDto {
  return {
    ...prevPrice,
    detectionTimestamp: formatISO(prevPrice.detectionTimestamp),
  };
}
export interface PreviousPriceDto {
  detectionTimestamp: string;
  price: number;
}
export function mapPrevPrice(prevPriceDto: PreviousPriceDto): PreviousPrice {
  return {
    ...prevPriceDto,
    detectionTimestamp: parseISO(prevPriceDto.detectionTimestamp),
  };
}

export interface Price {
  fueltype: FuelType;
  date: Date;
  price: number;
  prevPrices: PreviousPrice[];
}
export function mapPriceDto(price: Price): PriceDto {
  const prevPriceDtos = price.prevPrices.map(mapPrevPriceDto);
  let fueltype: FuelType | number = price.fueltype;
  if (typeof fueltype === "string") {
    fueltype = fuelTypeToInteger(fueltype);
  }
  return {
    ...price,
    fueltype: fueltype,
    date: formatISO(price.date),
    prevPrices: JSON.stringify(prevPriceDtos),
  };
}
export interface PriceDto {
  fueltype: number;
  date: string;
  price: number;
  prevPrices: string;
}
export function mapPrice(priceDto: PriceDto): Price {
  let prevPriceDtos: PreviousPriceDto[] = [];
  try {
    prevPriceDtos = JSON.parse(priceDto.prevPrices) as PreviousPriceDto[];
  } catch (error) {
    console.error("mapPrice failed", error, priceDto);
  }
  return {
    ...priceDto,
    fueltype: integerToFuelType(priceDto.fueltype),
    date: parseISO(priceDto.date),
    prevPrices: prevPriceDtos.map(mapPrevPrice),
  };
}

export interface DayPrices {
  today: Price | null;
  yesterday: Price | null;
  tomorrow: Price | null;
}

export interface LibsqlConnectionInfo {
  url: string;
  authToken: string;
}
export const getEnv = (): Env => {
  return {
    JOB_KEY: process.env.JOB_KEY!,
    LIBSQL_URL: process.env.LIBSQL_URL!,
    LIBSQL_AUTH_TOKEN: process.env.LIBSQL_AUTH_TOKEN!,
  };
};
