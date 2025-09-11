// references: https://www.blocknotejs.org/docs/features/blocks
// Prerequisite types from the general Blocknote structure
// This library is required for our editor app to read/write BlockNote types with
// explicit type definition.

export type Styles = {
  bold?: boolean;
  italic?: boolean;
  underline?: boolean;
  strike?: boolean;
  textColor?: string;
  backgroundColor?: string;
};

export type StyledText = {
  type: "text";
  text: string;
  styles: Styles;
};

export type Link = {
  type: "link";
  content: StyledText[];
  href: string;
};

export type InlineContent = StyledText | Link;

// Definition for the HeadingBlock

/**
 * Represents the properties specific to a Heading block.
 */
export type HeadingBlockProps = {
  /**
   * The heading level.
   */
  level: 1 | 2 | 3 | 4 | 5 | 6;
  /**
   * The text alignment.
   * @default "left"
   */
  textAlignment?: "left" | "center" | "right" | "justify";
  /**
   * The text color.
   * @default "default"
   */
  textColor?: string;
  /**
   * The background color.
   * @default "default"
   */
  backgroundColor?: string;
};

/**
 * Represents a Heading block in the editor.
 */
export type HeadingBlock = {
  id: string;
  type: "heading";
  props: HeadingBlockProps;
  content: InlineContent[];
  children: Block[]; // Assuming a generic 'Block' type for children
};

// A generic Block type for use in the `children` array
export type Block = {
  id: string;
  type: string;
  props: Record<string, any>;
  content: any;
  children: Block[];
};
