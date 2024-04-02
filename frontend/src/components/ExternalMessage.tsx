import React, { useState, useEffect } from "react";

interface Props {
  style: React.CSSProperties;
  data: string;
  textKey: string;
  fadeOut: boolean;
}

const ExternalMessage: React.FC<Props> = ({
  style,
  data,
  textKey,
  fadeOut,
}) => {
  return (
    <div
      className={`message external transition-opacity duration-500 ease-in-out ${
        fadeOut ? "opacity-0" : ""
      }`}
      id={textKey}
    >
      <p
        style={{
          ...style,
          position: "absolute",
          border: "none",
          background: "transparent",
          outline: "none",
          fontFamily: "Nunito, sans-serif",
          whiteSpace: "nowrap",
          fontSize: "0.875rem",
        }}
      >
        {data}
      </p>
    </div>
  );
};

export default ExternalMessage;
