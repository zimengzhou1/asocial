"use client";

import dynamic from "next/dynamic";

const Messages = dynamic(() => import("@/components/Messages"), {
  ssr: false,
});

export default function Chat() {
  return <Messages />;
}
