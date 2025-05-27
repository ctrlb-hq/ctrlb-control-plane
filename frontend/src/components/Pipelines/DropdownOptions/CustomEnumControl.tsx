import { withJsonFormsEnumProps } from '@jsonforms/react';
import {
    ControlProps
} from '@jsonforms/core';
import {
    isEnumControl,
    rankWith
} from '@jsonforms/core';
import {
    FormControl,
    InputLabel,
    MenuItem,
    Select
} from '@mui/material';

const CustomEnumControl = (props: ControlProps) => {
    const {
        data,
        id,
        enabled,
        schema,
        label,
        required,
        path,
        handleChange,
    } = props;

    const enumOptions = schema.enum || [];

    return (
        <FormControl fullWidth required={required} margin="normal">
  <InputLabel id={`${id}-label`}>
    {label}
  </InputLabel>
  <Select
    labelId={`${id}-label`}
    id={id}
    value={data ?? ''}
    onChange={(e) => handleChange(path, e.target.value)}
    disabled={!enabled}
    label={label}
    MenuProps={{
      disablePortal: true,
      container: document.body
    }}
  >
    {enumOptions.map((option: any, index: number) => (
      <MenuItem key={index} value={option}>
        {option}
      </MenuItem>
    ))}
  </Select>
</FormControl>

    );
};

export const customEnumRenderer = {
    tester: rankWith(5, isEnumControl),
    renderer: withJsonFormsEnumProps(CustomEnumControl)
};
export default CustomEnumControl;