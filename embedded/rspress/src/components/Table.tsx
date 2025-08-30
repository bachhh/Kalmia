import React from "react";
import { getColorClass, getStyles } from "../utils/style";
import { useDark } from "rspress/runtime";

interface DefaultProps {
  backgroundColor?: string;
  textColor?: string;
  textAlignment?: "left" | "center" | "right" | "justify";
}

interface TableProps {
  rawJson: TableData;
}

type TableData = {
  id: string;
  type: "table";
  props: DefaultProps;
  content: TableContent;
  children: Block[];
};

type TableContent = {
  type: "tableContent";
  columnWidths: number[];
  headerRows: number;
  rows: {
    cells: TableCell[];
  }[];
};

type TableCell = {
  type: "tableCell";
  props: {
    colspan?: number;
    rowspan?: number;
  };
  content: InlineContent[];
};

type Link = {
  type: "link";
  content: StyledText[];
  href: string;
};

type StyledText = {
  type: "text";
  text: string;
  styles: Styles;
};

type CustomInlineContent = {
  type: string;
  content: StyledText[] | undefined;
  props: Record<string, boolean | number | string>;
};

type InlineContent = Link | StyledText | CustomInlineContent;

interface Styles {
  bold?: boolean;
  textColor?: string;
  strike?: boolean;
  backgroundColor?: string;
  underline?: boolean;
}

export const Table: React.FC<TableProps> = ({ rawJson }) => {
  const { props, content } = rawJson;

  const containerClasses = [
    "kal-flex",
    "kal-justify-center",
    "kal-w-full",
    "kal-overflow-x-auto",
    "kal-py-4",
  ].join(" ");

  const tableClasses = [
    "kal-border-collapse",
    "kal-border",
    "kal-border-gray-300",
    getColorClass(props.textColor),
    getColorClass(props.backgroundColor, true),
    "kal-mx-auto",
    "kal-w-auto",
    "kal-max-w-full",
  ]
    .filter(Boolean)
    .join(" ");

  // styling for thead
  const headerClasses = [
    "kal-p-2",
    "kal-text-left",
    "kal-font-bold",
    "kal-uppercase",
    "kal-tracking-wider",
    "kal-text-gray-600",
    "kal-bg-gray-50",
    "kal-border",
    "kal-border-gray-300",
  ].join(" ");

  // TODO: adjust column width to BlockNote's param
  return (
    <div className={containerClasses}>
      <table className={tableClasses}>
        {/* if header row is set, promote the first row to <thead> */}
        {content.headerRows === 1 && (
          <thead>
            {content.rows.slice(0, 1).map((headerRow, rowIndex) => (
              <tr key={rowIndex}>
                {headerRow.cells.map((cell, cellIndex) => (
                  <th key={cellIndex} className={headerClasses}>
                    {cell.content.map((cellContent, cellContentIndex) =>
                      renderTableCell(cellContent, cellContentIndex),
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
        )}

        <tbody>
          {/* if header row is set, skip over the first row */}
          {content.rows
            .slice(content.headerRows === 1 ? 1 : 0)
            .map((row, rowIndex) => (
              <tr key={rowIndex}>
                {row.cells.map((cell, cellIndex) => (
                  <td
                    key={cellIndex}
                    className="kal-border kal-border-gray-300 kal-p-2"
                  >
                    {cell.content.map((cellContent, cellContentIndex) =>
                      renderTableCell(cellContent, cellContentIndex),
                    )}
                  </td>
                ))}
              </tr>
            ))}
        </tbody>
      </table>
    </div>
  );
};

const renderTableCell = (
  content: InlineContent,
  index: number,
): React.ReactNode => {
  // special case: inline image with preview on table cell
  // for now, we just have fixed name, always centered
  // container class and image class are hard-coded
  // TODO: use a proper CustomInlineContent
  if (content.type === "text") {
    const regex = /^img="(.+?)";previewWidth=(\d+)$/;
    const match = content.text.match(regex);

    if (match) {
      const url = match[1];
      const previewWidth = parseInt(match[2], 10);
      const imageStyle = { maxWidth: `${previewWidth}px` };

      const containerClasses = ["kal-items-center", "kal-p-4"].join(" ");

      const imageClasses = [
        "kal-max-w-full",
        "kal-h-auto",
        "kal-object-contain",
      ]
        .filter(Boolean)
        .join(" ");

      return (
        <div key={index} className={containerClasses}>
          <img
            src={url}
            alt="previewImage"
            className={imageClasses}
            style={imageStyle}
          />
        </div>
      );
    }
  }

  // type Link
  if (content.type === "link") {
    const linkContent = content.content[0]; // assuming Link type only have a single block of content
    const linkClass = [
      ...getStyles(linkContent.styles),
      "kal-underline",
      "kal-text-blue-500",
    ];

    return (
      <a
        key={index}
        href={content.href || ""}
        className={linkClass.join(" ")}
        target="_blank"
      >
        {linkContent.text?.trim() || ""}
      </a>
    );
  }

  // StyledText
  return (
    <span key={index} className={getStyles(content.styles).join(" ")}>
      {content.text}
    </span>
  );
};

const getStyles = (styles: CellStyles = {}) =>
  [
    styles.bold ? "kal-font-bold" : "",
    getColorClass(styles.textColor),
    getColorClass(styles.backgroundColor, true),
    styles.strike ? "kal-line-through" : "",
    styles.underline ? "kal-underline" : "",
  ].filter(Boolean);

export default Table;
