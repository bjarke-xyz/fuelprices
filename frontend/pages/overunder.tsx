import { ChartData } from "chart.js";
import { format, parse, parseISO, subDays } from "date-fns";
import { da } from "date-fns/locale";
import { useEffect, useMemo, useState } from "react";
import { Line } from "react-chartjs-2";
import useSWR from "swr";
import fetcher from "../utils/fetcher";
import { Fueltype } from "../utils/types";
import { HistoricalPriceResult } from "./historical";
const dateFormat = "yyyy-MM-dd";

interface GuessOutcome {
  guess: ButtonType;
  currentPrice: number;
  currentDate: Date;
  nextPrice: number;
  nextDate: Date;
}

type ButtonType = "over" | "under";

export const OverUnder: React.FC<{ fueltype: Fueltype }> = ({ fueltype }) => {
  const refDate = useMemo(() => new Date(0), []);
  const [staticNow] = useState<Date>(new Date());
  const [from, setFrom] = useState<Date>(subDays(staticNow, 31));
  const [inputFrom, setInputFrom] = useState<string>(
    format(from, "yyyy-MM-dd")
  );
  const [visibleDataIndex, setVisibleDataIndex] = useState<number>(1);
  const [visibleData, setVisibleData] = useState<HistoricalPriceResult>({
    dates: [],
    prices: [],
  });
  const [guesses, setGuesses] = useState<GuessOutcome[]>([]);
  const { data, error } = useSWR<HistoricalPriceResult>(
    `/api/price-trends?type=${fueltype}&from=${format(
      from,
      "yyyy-MM-dd"
    )}&to=${format(staticNow, "yyyy-MM-dd")}`,
    fetcher
  );
  if (error) {
    console.log(error);
  }
  useEffect(() => {
    if (inputFrom) {
      const newFrom = parse(inputFrom, "yyyy-MM-dd", refDate);
      if (newFrom !== refDate) {
        setFrom(newFrom);
      }
    }
  }, [inputFrom, refDate]);
  useEffect(() => {
    setVisibleDataIndex(1);
    setGuesses([]);
  }, [from]);
  useEffect(() => {
    if (!data) return;
    setVisibleData({
      dates: data.dates.slice(0, visibleDataIndex),
      prices: data.prices.slice(0, visibleDataIndex),
    });
  }, [data, visibleDataIndex]);
  const chartData: ChartData<"line", number[], string> = {
    labels: visibleData.dates.map((dateStr) => {
      const date = parseISO(dateStr);
      return format(date, "yyyy-MM-dd");
    }),
    datasets: [
      {
        label: fueltype,
        data: visibleData.prices,
        borderColor: "rgb(255, 99, 132)",
        backgroundColor: "rgba(255, 99, 132, 0.5)",
      },
    ],
  };
  const chartOptions = {
    maintainAspectRatio: false,
    scales: {
      y: {
        ticks: {
          autoSkip: true,
          // Mannipulate value by function
          // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unused-vars
          callback: function (value: any, _index: any, _values: any) {
            return value.toFixed(2); // Get rounding to 2 decimal places
          },
        },
      },
    },
  };

  function handleButtonClick(btn: ButtonType) {
    if (!data || visibleData.prices.length === 0) return;
    const currentPrice = visibleData.prices.at(-1);
    const currentDateStr = visibleData.dates.at(-1);
    const currentDate = parseISO(currentDateStr!);
    const nextPrice = data.prices[visibleDataIndex];
    const nextDateStr = data.dates[visibleDataIndex];
    const nextDate = parseISO(nextDateStr);
    if (!nextPrice || !currentPrice) return;
    setGuesses((currentGuesses) => [
      { currentPrice, currentDate, nextPrice, nextDate, guess: btn },
      ...currentGuesses,
    ]);
    setVisibleDataIndex((curr) => curr + 1);
  }
  return (
    <div className="flex flex-col">
      <div>
        <form className="flex flex-row justify-evenly">
          <div>
            <input
              id="from"
              value={inputFrom}
              onChange={(e) => setInputFrom(e.target.value)}
              type="date"
              className="bg-white text-slate-700 border border-solid border-slate-300 rounded transition"
            ></input>
          </div>
        </form>
      </div>
      <div>
        <Line data={chartData} options={chartOptions} height="300px" />
      </div>
      <div className="text-center text-lg p-2">
        {data &&
        visibleData.prices.length > 0 &&
        visibleData.dates.length > 0 ? (
          <>
            <p>
              Prisen var {visibleData.prices.at(-1)} kr.{" "}
              {format(parseISO(visibleData.dates.at(-1)!), "PPPP", {
                locale: da,
              })}
            </p>
            {visibleDataIndex + 1 < data.prices.length ? (
              <p>Var prisen under eller over dagen efter?</p>
            ) : null}
          </>
        ) : null}
      </div>
      <div className="flex flex-row gap-4 justify-center items-center flex-wrap">
        <OverUnderButton
          text="Under"
          disabled={!data || visibleDataIndex + 1 >= data.prices.length}
          onClick={() => handleButtonClick("under")}
        />
        <OverUnderButton
          text="Over"
          disabled={!data || visibleDataIndex + 1 >= data.prices.length}
          onClick={() => handleButtonClick("over")}
        />
      </div>
      {guesses.length > 0 ? (
        <div className="flex justify-center mt-4 mx-2">
          <OverUnderHistory guesses={guesses} />
        </div>
      ) : null}
    </div>
  );
};

interface OverUnderHistoryProps {
  guesses: GuessOutcome[];
}
const OverUnderHistory: React.FC<OverUnderHistoryProps> = ({ guesses }) => {
  return (
    <table className="table-auto border-spacing-2 border-spacing-y-4 border border-slate-400 border-separate rounded">
      <tbody>
        {guesses.map((guess, i) => (
          <OverUnderHistoryRow key={i} guess={guess} />
        ))}
      </tbody>
    </table>
  );
};

const OverUnderHistoryRow: React.FC<{ guess: GuessOutcome }> = ({ guess }) => {
  function getReaction(): "correct" | "incorrect" | "na" {
    if (
      (guess.guess === "over" && guess.nextPrice > guess.currentPrice) ||
      (guess.guess === "under" && guess.nextPrice < guess.currentPrice)
    ) {
      return "correct";
    }
    if (
      (guess.guess === "over" && guess.nextPrice < guess.currentPrice) ||
      (guess.guess === "under" && guess.nextPrice > guess.currentPrice)
    ) {
      return "incorrect";
    }
    return "na";
  }
  function getReactionEmoji(reaction: "correct" | "incorrect" | "na"): string {
    if (reaction === "correct") return "✅";
    if (reaction === "incorrect") return "❌";
    if (reaction === "na") return "😐";
    return "";
  }

  function getCompare(): string {
    if (guess.currentPrice > guess.nextPrice) return "⬇️";
    if (guess.currentPrice < guess.nextPrice) return "⬆️";
    return "🟰";
  }
  const reaction = getReaction();
  const reactionEmoji = getReactionEmoji(reaction);
  const compare = getCompare();
  return (
    <tr>
      <td>
        {guess.currentPrice} kr. ({format(guess.currentDate, dateFormat)})
      </td>
      <td>{compare}</td>
      <td>
        {guess.nextPrice} kr. ({format(guess.nextDate, dateFormat)})
      </td>
      <td>{guess.guess}</td>
      <td
        className={`p-4 rounded-sm ${
          reaction === "correct"
            ? "bg-green-200"
            : reaction === "incorrect"
            ? "bg-red-200"
            : reaction === "na"
            ? "bg-gray-300"
            : null
        }`}
      >
        {reactionEmoji}
      </td>
    </tr>
  );
};

interface OverUnderButtonProps {
  text: string;
  onClick: () => unknown;
  disabled: boolean;
}
const OverUnderButton: React.FC<OverUnderButtonProps> = ({
  text,
  onClick,
  disabled,
}) => {
  return (
    <button
      disabled={disabled}
      onClick={onClick}
      className="bg-emerald-200 hover:bg-emerald-300 disabled:bg-gray-300 dark:text-slate-900 w-32 h-32 rounded shadow disabled:shadow-none"
    >
      {text}
    </button>
  );
};
