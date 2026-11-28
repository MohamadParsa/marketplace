import { Add } from "@mui/icons-material";
import { Button, type ButtonProps } from "@mui/material";

export function CreateButton(props: ButtonProps) {
  return (
    <Button
      {...props}
      variant="contained"
      color="primary"
      size="medium"
      disableElevation
      startIcon={<Add />}
      className={["admin-create-button", props.className]
        .filter(Boolean)
        .join(" ")}
    />
  );
}
