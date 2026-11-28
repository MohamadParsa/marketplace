import { Component, type ReactNode } from "react";
import { Alert, Button } from "@mui/material";

// Keep a failed page inside the panel; navigation and the session stay mounted.
export class PageErrorBoundary extends Component<
  { children: ReactNode },
  { failed: boolean }
> {
  state = { failed: false };

  static getDerivedStateFromError() {
    return { failed: true };
  }

  render() {
    if (this.state.failed)
      return (
        <Alert severity="error">
          نمایش این صفحه با خطا روبه‌رو شد. می‌توانید دوباره تلاش کنید یا از منوی
          پنل به صفحهٔ دیگری بروید.
          <Button onClick={() => this.setState({ failed: false })}>
            تلاش دوباره
          </Button>
        </Alert>
      );
    return this.props.children;
  }
}
