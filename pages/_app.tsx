import "../styles/globals.css";
import type { AppProps } from "next/app";
import { Header } from "../components/header";
import { Footer } from "../components/footer";
import { useState } from "react";

export type Fueltype = "unleaded95" | "diesel" | "octane100";

function MyApp({ Component, pageProps }: AppProps) {
  const [fueltype, setFueltype] = useState<Fueltype>("unleaded95");
  return (
    <div className="h-screen flex flex-col">
      <Header fueltype={fueltype} setFueltype={setFueltype} />
      <main className="mb-auto mt-10">
        <Component {...pageProps} fueltype={fueltype} />
      </main>
      <Footer />
    </div>
  );
}

export default MyApp;
