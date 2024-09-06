import Link from "next/link";
import { useRouter } from "next/router";
import { ChangeEvent } from "react";
import { Fueltype } from "../pages/_app";

interface HeaderProps {
  fueltype: Fueltype;
  setFueltype: (fueltype: Fueltype) => void;
}

export const Header: React.FC<HeaderProps> = ({ fueltype, setFueltype }) => {
  const router = useRouter();
  const currentRoute = router.pathname;

  const setFueltypeHelper = (e: ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value as Fueltype;
    setFueltype(value);
  };

  return (
    <div className="flex flex-col items-center py-2">
      <div>
        <span className="text-2xl">Fuel Prices</span>{" "}
      </div>
      <div>
        <form>
          <select
            value={fueltype}
            onChange={setFueltypeHelper}
            className="bg-white text-slate-700 border border-solid border-slate-300 rounded transition"
          >
            <option value="octane95">Octane 95</option>
            <option value="diesel">Diesel</option>
            <option value="octane100">Octane 100</option>
          </select>
        </form>
      </div>
      <div className="flex flex-row space-x-4">
        <div>
          <Link href="/">
            <a className={currentRoute === "/" ? "font-bold" : ""}>Today</a>
          </Link>
        </div>
        <div>
          <Link href="/historical">
            <a className={currentRoute === "/historical" ? "font-bold" : ""}>
              Historical
            </a>
          </Link>
        </div>
        <div>
          <Link href="/overunder">
            <a className={currentRoute === "/overunder" ? "font-bold" : ""}>
              Over - Under
            </a>
          </Link>
        </div>
      </div>
    </div>
  );
};
