import React, { useState, useEffect } from "react";

interface Props {
  style: React.CSSProperties;
  data: string;
  timeoutID: number;
  textKey: string;
  fadeOut: boolean;
  onInputChange: (textKey: string, data: string) => void;
}

const LocalInput: React.FC<Props> = ({
  style,
  data,
  timeoutID,
  textKey,
  fadeOut,
  onInputChange,
}) => {
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onInputChange(textKey, e.target.value);
  };

  return (
    <div
      className={`message external transition-opacity duration-500 ease-in-out ${
        fadeOut ? "opacity-0" : ""
      }`}
      id={textKey}
    >
      <input
        id={textKey}
        value={data}
        onChange={handleInputChange}
        style={{
          ...style,
          position: "absolute",
          border: "none",
          background: "transparent",
          outline: "none",
          fontFamily: "Nunito, sans-serif",
          width: "450px",
          overflow: "auto",
          fontSize: "0.875rem",
        }}
        autoFocus
        maxLength={60}
      />
    </div>
  );
};

export default LocalInput;
