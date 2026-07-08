"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import { operatorAssignmentsApi, branchesApi, countersApi } from "~/lib/api/qms";
import { usersApi } from "~/lib/api/users";
import type { OperatorAssignmentResponse } from "@casbin/api-types";

interface UserOption { id: string; name: string; }
interface BranchOption { id: string; code: string; }
interface CounterOption { id: string; code: string; }

export function OperatorAssignmentsContent() {
  const [assignments, setAssignments] = useState<OperatorAssignmentResponse[]>([]);
  const [branches, setBranches] = useState<BranchOption[]>([]);
  const [users, setUsers] = useState<UserOption[]>([]);
  const [counters, setCounters] = useState<CounterOption[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedBranch, setSelectedBranch] = useState("");
  const [selectedUser, setSelectedUser] = useState("");
  const [selectedCounter, setSelectedCounter] = useState("");

  const load = useCallback(async () => {
    try {
      const [assignResp, branchResp, counterResp, userResp] = await Promise.all([
        operatorAssignmentsApi.getAll(),
        branchesApi.getAll(),
        countersApi.getAll(),
        usersApi.getAll(),
      ]);
      setAssignments(assignResp.data || []);
      setBranches((branchResp.data || []).map((b: any) => ({ id: b.id, code: b.code || b.name })));
      setCounters((counterResp.data || []).map((c: any) => ({ id: c.id, code: c.code || c.name })));
      setUsers((userResp.data || []).map((u: any) => ({ id: u.id, name: u.name || u.email })));
    } catch (error: any) {
      toast.error(error.message || "Failed to load");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleCreate = async () => {
    if (!selectedBranch || !selectedUser || !selectedCounter) {
      toast.error("Branch, user, and counter required");
      return;
    }
    try {
      await operatorAssignmentsApi.create({ branch_id: selectedBranch, user_id: selectedUser, counter_id: selectedCounter });
      toast.success("Operator assigned");
      setSelectedBranch("");
      setSelectedUser("");
      setSelectedCounter("");
      load();
    } catch (error: any) {
      toast.error(error.message || "Failed to create");
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await operatorAssignmentsApi.delete(id);
      toast.success("Assignment removed");
      load();
    } catch (error: any) {
      toast.error(error.message || "Failed to delete");
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold tracking-tight">Operator Assignments</h2>
      <p className="text-muted-foreground">Assign operators to counters.</p>

      <Card>
        <CardHeader><CardTitle>New Assignment</CardTitle><CardDescription>Select branch, user, and counter.</CardDescription></CardHeader>
        <CardContent className="flex flex-wrap items-end gap-3">
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium">Branch</label>
            <Select value={selectedBranch} onValueChange={setSelectedBranch}>
              <SelectTrigger className="w-44"><SelectValue placeholder="Branch..." /></SelectTrigger>
              <SelectContent>{branches.map((b) => <SelectItem key={b.id} value={b.id}>{b.code}</SelectItem>)}</SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium">User</label>
            <Select value={selectedUser} onValueChange={setSelectedUser}>
              <SelectTrigger className="w-44"><SelectValue placeholder="User..." /></SelectTrigger>
              <SelectContent>{users.map((u) => <SelectItem key={u.id} value={u.id}>{u.name}</SelectItem>)}</SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium">Counter</label>
            <Select value={selectedCounter} onValueChange={setSelectedCounter}>
              <SelectTrigger className="w-44"><SelectValue placeholder="Counter..." /></SelectTrigger>
              <SelectContent>{counters.map((c) => <SelectItem key={c.id} value={c.id}>{c.code}</SelectItem>)}</SelectContent>
            </Select>
          </div>
          <Button onClick={handleCreate}><Icon name="Plus" className="mr-1 h-4 w-4" />Assign</Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Current Assignments</CardTitle></CardHeader>
        <CardContent>{isLoading ? <p className="text-muted-foreground">Loading...</p>
          : assignments.length === 0 ? <p className="text-muted-foreground">No assignments yet.</p>
          : <Table><TableHeader><TableRow><TableHead>User</TableHead><TableHead>Branch</TableHead><TableHead>Counter</TableHead><TableHead>Assigned At</TableHead><TableHead className="w-20">Action</TableHead></TableRow></TableHeader><TableBody>{assignments.map((a) => (<TableRow key={a.id}><TableCell>{a.user_id}</TableCell><TableCell>{a.branch_id}</TableCell><TableCell>{a.counter_id}</TableCell><TableCell>{new Date(a.assigned_at).toLocaleString()}</TableCell><TableCell><Button variant="destructive" size="sm" onClick={() => handleDelete(a.id)}>Remove</Button></TableCell></TableRow>))}</TableBody></Table>}</CardContent>
      </Card>
    </div>
  );
}
