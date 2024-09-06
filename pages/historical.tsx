import {
  CategoryScale,
  Chart as ChartJS,
  ChartData,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip,
} from "chart.js";
import { format, parse, parseISO, subDays } from "date-fns";
import { NextPage } from "next";
import { useEffect, useState } from "react";
import { Line } from "react-chartjs-2";
import useSWR from "swr";
import { API_URL } from "../utils/constants";
import { Fueltype } from "./_app";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

export interface HistoricalPriceResult {
  dates: string[];
  prices: number[];
}

export const fetcher = (
  url: string,
  fueltype: Fueltype,
  from: Date,
  to: Date
) =>
  fetch(
    `/api/price-trends?type=${fueltype}&from=${format(
      from,
      "yyyy-MM-dd"
    )}&to=${format(to, "yyyy-MM-dd")}`
  ).then((res) => res.json());
const Historical: NextPage<{ fueltype: Fueltype }> = ({ fueltype }) => {
  const refDate = new Date(0);
  const staticNow = new Date();
  const [from, setFrom] = useState<Date>(subDays(staticNow, 365));
  const [inputFrom, setInputFrom] = useState<string>(
    format(from, "yyyy-MM-dd")
  );
  const [to, setTo] = useState<Date>(new Date());
  const [inputTo, setInputTo] = useState<string>(format(to, "yyyy-MM-dd"));
  const { data, error } = useSWR<HistoricalPriceResult>(
    [API_URL, fueltype, from, to],
    // FIXME: any use. this started giving an error after @cloudflare/workers-types was added to tsconfig
    fetcher as any
  );

  useEffect(() => {
    if (inputFrom) {
      const newFrom = parse(inputFrom, "yyyy-MM-dd", refDate);
      if (newFrom !== refDate) {
        setFrom(newFrom);
      }
    }
  }, [inputFrom]);
  useEffect(() => {
    if (inputTo) {
      const newTo = parse(inputTo, "yyyy-MM-dd", refDate);
      if (newTo !== refDate) {
        setTo(newTo);
      }
    }
  }, [inputTo]);

  const chartData: ChartData<"line", number[], string> = {
    labels: ((data ?? {})?.dates ?? []).map((dateStr) => {
      const date = parseISO(dateStr);
      return format(date, "yyyy-MM-dd");
    }),
    datasets: [
      {
        label: fueltype,
        data: ((data ?? {})?.prices ?? []),
        borderColor: "rgb(255, 99, 132)",
        backgroundColor: "rgba(255, 99, 132, 0.5)",
      },
    ],
  };

  return (
    <div className="flex flex-col">
      <div>
        <form>
          <div className="flex flex-row justify-evenly">
            <div className="border p-2 rounded">
              <label className="block" htmlFor="to">
                To
              </label>
              <input
                id="To"
                value={inputFrom}
                onChange={(e) => setInputFrom(e.target.value)}
                type="date"
                className="bg-white text-slate-700 border border-solid border-slate-300 rounded transition"
              ></input>
            </div>
            <div className="border p-2 rounded">
              <label className="block" htmlFor="from">
                From
              </label>
              <input
                id="from"
                value={inputTo}
                onChange={(e) => setInputTo(e.target.value)}
                type="date"
                className="bg-white text-slate-700 border border-solid border-slate-300 rounded transition"
              ></input>
            </div>
          </div>
        </form>
      </div>
      <div>
        {error && <div>An error has occurred</div>}
        {!data && <div>Loading...</div>}
        {!error && data && <Line data={chartData} />}
      </div>
    </div>
  );
};

export default Historical;
