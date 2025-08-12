import React from "react";

export default function FinalRoundBoard({
  availableCategories,
  canRemoveCategory,
  removeCategory,
}: {
  availableCategories: { [key: string]: boolean };
  canRemoveCategory: boolean;
  removeCategory: (category: string) => void;
}) {
  return (
    <ul className="w-full h-full">
      { Object.keys(availableCategories).map((category, index) => 
        <li key={index} onClick={() => {
          if (canRemoveCategory) removeCategory(category);
        }}>{category}</li>
      ) }
    </ul>
  );
}
