import { JSX } from "preact/jsx-runtime";

interface GenericTableProps<T> {
  data: T[];
  columns: { header: string; accessor: (item: T) => JSX.Element }[];
}

export default function GenericTable<T>({ data, columns }: GenericTableProps<T>) {
  return (
    <table class="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400">
      <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
        <tr>
          {columns.map((col) => (
            <th scope="col" class="px-6 py-3">
              {col.header}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {data.map((item, index) => (
          <tr key={index} class="bg-white border-b dark:bg-gray-800 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600">
            {columns.map((col, colIndex) => (
              <td key={colIndex} class="px-6 py-4">
                {col.accessor(item)}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
