import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";

export function ProductThumbnail({
  imageID,
  name,
}: {
  imageID?: string;
  name: string;
}) {
  const { loadCatalogMedia, authToken } = useWorkspace();
  const [src, setSrc] = useState("");
  const query = useQuery({
    queryKey: ["product-thumbnail", authToken, imageID],
    queryFn: () => loadCatalogMedia(imageID!),
    enabled: !!imageID,
    staleTime: 60000,
  });
  useEffect(() => {
    if (!query.data) {
      setSrc("");
      return;
    }
    const url = URL.createObjectURL(query.data);
    setSrc(url);
    return () => URL.revokeObjectURL(url);
  }, [query.data]);
  return src ? (
    <img className="product-grid-thumbnail" src={src} alt={name} />
  ) : (
    <span
      className="product-grid-thumbnail product-grid-thumbnail-empty"
      aria-label="بدون تصویر"
    >
      —
    </span>
  );
}
