
interface CheckboxProps {
  checked: boolean;
  onInput: () => void;
}

export default function Checkbox({ checked, onInput }: CheckboxProps) {
  return (
    <div class="flex items-center">
      <input
        id="checkbox-all-search"
        type="checkbox"
        class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 dark:focus:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
        onInput={onInput}
        checked={checked}
      />
      <label for="checkbox-all-search" class="sr-only">
        checkbox
      </label>
    </div>
  );
}
