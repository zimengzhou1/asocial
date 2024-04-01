import React, { useState, useEffect } from "react";

interface Props {
  style: React.CSSProperties;
  data: string;
  timeoutID: number;
  textKey: string;
  onInputChange: (textKey: string, data: string) => void;
}

const LocalInput: React.FC<Props> = ({
  style,
  data,
  timeoutID,
  textKey,
  onInputChange,
}) => {
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onInputChange(textKey, e.target.value);
  };

  return (
    <div id={textKey}>
      <input
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
