SELECT s.Id AS SupplierId,p.SupplyDelay,p.SupplyDelayPeriodId FROM Supplier s LEFT JOIN SupplierProfile p ON p.Id=s.SupplierProfileId WHERE s.IsActive=1 AND s.Deleted=0
