"use client";

import dynamic from "next/dynamic";
import { use } from "react";

const Messages = dynamic(() => import("@/components/Messages"), {
  ssr: false,
});

interface ChatPageProps {
  params: Promise<{ slug: string }>;
}

export default function Chat({ params }: ChatPageProps) {
  const { slug } = use(params);
  return <Messages roomSlug={slug} />;
}
