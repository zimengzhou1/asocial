import Link from "next/link";
import React from "react";

const MyComponent: React.FC = () => {
  return (
    <ul className="space-y-1">
      <li>
        <details className="group [&_summary::-webkit-details-marker]:hidden">
          <summary className="flex cursor-pointer items-center justify-between rounded-lg px-4 py-2">
            <span className=" font-custom mr-2"> asocialpage </span>

            <span className="shrink-0 transition duration-300 group-open:-rotate-180">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-5 w-5"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                  clipRule="evenodd"
                />
              </svg>
            </span>
          </summary>

          <ul className="">
            <li>
              <Link
                href="/"
                className="block rounded-lg px-4 py-1 text-sm font-custom hover:font-bold"
              >
                home
              </Link>
            </li>

            <li>
              <a
                href="#"
                className="block rounded-lg px-4 py-1 text-sm font-custom hover:font-bold"
              >
                feedback
              </a>
            </li>
            <li>
              <a
                href="#"
                className="block rounded-lg px-4 py-1 text-sm font-custom hover:font-bold"
              >
                about
              </a>
            </li>
          </ul>
        </details>
      </li>
    </ul>
  );
};

export default MyComponent;
