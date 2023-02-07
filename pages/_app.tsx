import type { AppProps } from "next/app";
import { CSSProperties, useEffect, useState } from "react";
import { Footer } from "../components/footer";
import { Header } from "../components/header";
import { ThemeProvider } from "../hooks/theme-context";
import "../styles/globals.css";

export type Fueltype = "unleaded95" | "diesel" | "octane100";

export interface Appearance {
  bgColor: string | null;
  textColor: string | null;
}

function MyApp({ Component, pageProps }: AppProps) {
  const [fueltype, setFueltype] = useState<Fueltype>("unleaded95");
  const [showing, setShowing] = useState(false);

  // nextjs hacks https://stackoverflow.com/a/71797054
  useEffect(() => {
    setShowing(true);
  }, []);

  if (!showing) {
    return null;
  }
  if (typeof window === "undefined") {
    return <></>;
  } else {
    const searchParams = new URLSearchParams(window.location.search);
    const bgColor = searchParams.get("bgColor");
    const textColor = searchParams.get("textColor");
    let colorClasses = "";
    const inlineStyle: CSSProperties = {};
    if (!bgColor) {
      colorClasses += "bg-white dark:bg-slate-800 ";
    } else {
      inlineStyle.backgroundColor = `#${bgColor}`;
    }
    if (!textColor) {
      colorClasses += "text-slate-900 dark:text-white";
    } else {
      inlineStyle.color = `#${textColor}`;
    }
    const appearance: Appearance = { bgColor, textColor };
    return (
      <ThemeProvider>
        <div
          className={`h-screen flex flex-col ${colorClasses}`}
          style={inlineStyle}
        >
          <Header fueltype={fueltype} setFueltype={setFueltype} />
          <main className="mb-auto mt-10">
            <Component
              {...pageProps}
              fueltype={fueltype}
              appearance={appearance}
            />
          </main>
          <Footer />
        </div>
      </ThemeProvider>
    );
  }
}

export default MyApp;
