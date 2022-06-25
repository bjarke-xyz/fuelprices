import { Head, Html, Main, NextScript } from "next/document";
import { useEffect } from "react";

export default function Document() {
  return (
    <Html lang="en">
      <Head />
      <body className="bg-white text-slate-900 dark:bg-slate-800 dark:text-white">
        <Main />
        <NextScript />
      </body>
    </Html>
  );
}
